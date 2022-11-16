package main

import (
	"context"
	"github.com/gin-gonic/gin"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/controller"
	"github.com/konveyor/tackle2-hub/database"
	"github.com/konveyor/tackle2-hub/importer"
	"github.com/konveyor/tackle2-hub/k8s"
	crd "github.com/konveyor/tackle2-hub/k8s/api"
	"github.com/konveyor/tackle2-hub/migration"
	"github.com/konveyor/tackle2-hub/reaper"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
	"github.com/konveyor/tackle2-hub/volume"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"syscall"
)

var Settings = &settings.Settings

var log = logging.WithName("hub")

func init() {
	_ = Settings.Load()
}

//
// Setup the DB and models.
func Setup() (db *gorm.DB, err error) {
	err = migration.Migrate(migration.All())
	if err != nil {
		return
	}
	db, err = database.Open(true)
	if err != nil {
		return
	}
	return
}

//
// buildScheme adds CRDs to the k8s scheme.
func buildScheme() (err error) {
	err = crd.AddToScheme(scheme.Scheme)
	return
}

//
// addonManager
func addonManager(db *gorm.DB, adminChanged chan int) (mgr manager.Manager, err error) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		_ = http.ListenAndServe(":2112", nil)
	}()
	cfg, err := config.GetConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	mgr, err = manager.New(
		cfg,
		manager.Options{
			MetricsBindAddress: Settings.Metrics.Address(),
			Namespace:          Settings.Hub.Namespace,
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = controller.Add(mgr, db, adminChanged)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// main.
func main() {
	log.Info("Started", "settings", Settings)
	var err error
	defer func() {
		if err != nil {
			log.Trace(err)
		}
	}()
	syscall.Umask(0)
	//
	// Model
	db, err := Setup()
	if err != nil {
		panic(err)
	}
	//
	// k8s scheme.
	err = buildScheme()
	if err != nil {
		return
	}
	//
	// Add controller.
	adminChanged := make(chan int, 1)
	addonManager, err := addonManager(db, adminChanged)
	if err != nil {
		return
	}
	go func() {
		err = addonManager.Start(make(<-chan struct{}))
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}()
	//
	// k8s client.
	client, err := k8s.NewClient()
	if err != nil {
		err = liberr.Wrap(err)
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
	// Volumes.
	volumeManager := volume.Manager{
		Client: client,
		DB:     db,
	}
	volumeManager.Run(adminChanged)
	//
	// Application import.
	importManager := importer.Manager{
		DB: db,
	}
	importManager.Run(context.Background())
	//
	// Web
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	for _, h := range api.All() {
		h.With(db, client)
		h.AddRoutes(router)
	}
	err = router.Run()
}
