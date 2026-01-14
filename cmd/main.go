package main

import (
	"context"
	"net/http"
	"runtime/debug"
	"syscall"

	"github.com/gin-gonic/gin"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/api"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/internal/controller"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/heap"
	"github.com/konveyor/tackle2-hub/internal/importer"
	"github.com/konveyor/tackle2-hub/internal/k8s"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api"
	"github.com/konveyor/tackle2-hub/internal/metrics"
	"github.com/konveyor/tackle2-hub/internal/migration"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/reaper"
	"github.com/konveyor/tackle2-hub/internal/seed"
	"github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/internal/tracker"
	"github.com/konveyor/tackle2-hub/shared/command"
	"github.com/konveyor/tackle2-hub/shared/scm"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"github.com/konveyor/tackle2-hub/shared/ssh"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	Settings = &settings.Settings
	Log      = logr.New("hub", 0)
)

func init() {
	command.Log = logr.New("command", Settings.Log.Command)
	scm.Log = logr.New("scm", Settings.Log.SCM)
	ssh.Log = logr.New("ssh", Settings.Log.SSH)
}

// Setup the DB and models.
func Setup() (db *gorm.DB, err error) {
	err = migration.Migrate(migration.All())
	if err != nil {
		return
	}
	err = seed.Seed()
	if err != nil {
		return
	}
	db, err = database.Open(true)
	if err != nil {
		return
	}
	err = database.PK.Load(db, model.ALL)
	if err != nil {
		return
	}
	return
}

// buildScheme adds CRDs to the k8s scheme.
func buildScheme() (err error) {
	err = crd.AddToScheme(scheme.Scheme)
	return
}

// addonManager
func addonManager(db *gorm.DB) (mgr manager.Manager, err error) {
	cfg, err := config.GetConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	mgr, err = manager.New(
		cfg,
		manager.Options{
			MetricsBindAddress: "0",
			Namespace:          Settings.Hub.Namespace,
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = controller.Add(mgr, db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// main.
func main() {
	Log.Info("Started:\n" + Settings.String())
	var err error
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	syscall.Umask(0)
	debug.SetGCPercent(20)
	heap.Monitor()
	//
	// Model
	db, err := Setup()
	if err != nil {
		panic(err)
	}
	if !Settings.Disconnected {
		//
		// k8s scheme.
		err = buildScheme()
		if err != nil {
			return
		}
		//
		// Add controller.
		addonManager, aErr := addonManager(db)
		if aErr != nil {
			err = aErr
			return
		}
		go func() {
			err = addonManager.Start(context.Background())
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}()
	}
	//
	// k8s client.
	client, err := k8s.NewClient()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	// Document migration.
	jsdMigrator := migration.DocumentMigrator{
		DB:     db,
		Client: client,
	}
	err = jsdMigrator.Migrate(model.ALL)
	if err != nil {
		return
	}
	//
	// Auth
	if settings.Settings.Auth.Required {
		r := auth.NewReconciler(
			settings.Settings.Auth.Keycloak.Host,
			settings.Settings.Auth.Keycloak.Realm,
			settings.Settings.Auth.Keycloak.ClientID,
			settings.Settings.Auth.Keycloak.ClientSecret,
			settings.Settings.Auth.Keycloak.Admin.User,
			settings.Settings.Auth.Keycloak.Admin.Pass,
			settings.Settings.Auth.Keycloak.Admin.Realm,
		)
		err = r.Reconcile()
		if err != nil {
			return
		}
		auth.Hub = &auth.Builtin{}
		auth.Remote = auth.NewKeycloak(
			settings.Settings.Auth.Keycloak.Host,
			settings.Settings.Auth.Keycloak.Realm,
		)
	}
	//
	// Task
	taskManager := task.Manager{
		Client: client,
		DB:     db,
	}
	taskManager.Run(context.Background())
	//
	// Reaper
	reaperManager := reaper.Manager{
		Client: client,
		DB:     db,
	}
	reaperManager.Run(context.Background())
	//
	// Application import.
	importManager := importer.Manager{
		DB:          db,
		TaskManager: &taskManager,
		Client:      client,
	}
	importManager.Run(context.Background())
	//
	// Ticket trackers.
	trackerManager := tracker.Manager{
		DB: db,
	}
	trackerManager.Run(context.Background())
	//
	// Metrics
	if Settings.Metrics.Enabled {
		Log.Info("Serving Prometheus metrics", "port", Settings.Metrics.Port)
		http.Handle("/metrics", api.MetricsHandler())
		go func() {
			_ = http.ListenAndServe(Settings.Metrics.Address(), nil)
		}()
		metricsManager := metrics.Manager{
			DB: db,
		}
		metricsManager.Run(context.Background())
	}
	// Web
	router := gin.Default()
	router.Use(
		func(ctx *gin.Context) {
			rtx := api.RichContext(ctx)
			rtx.TaskManager = &taskManager
			rtx.DB = db
			rtx.Client = client
			ctx.Next()
			rtx.Detach()
		})
	router.Use(api.Render())
	router.Use(api.ErrorHandler())
	for _, h := range api.All() {
		h.AddRoutes(router)
	}
	err = router.Run()
}
