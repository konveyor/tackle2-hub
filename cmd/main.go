package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/api"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/internal/controller"
	"github.com/konveyor/tackle2-hub/internal/database"
	"github.com/konveyor/tackle2-hub/internal/frontend"
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

// port returns the API port.
func port() (port string) {
	port = fmt.Sprintf(":%d", Settings.API.Port)
	return
}

// Run router.
// The /hub (ingress) prefix is stripped.
func Run(e *gin.Engine) (err error) {
	err = http.ListenAndServe(
		port(),
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, "/hub")
				e.ServeHTTP(w, r)
			}))
	return
}

// main.
// Note: The initialization order is very important.
func main() {
	Log.Info("Started:\n" + Settings.String())
	ctx := context.Background()
	var err error
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	syscall.Umask(0)
	debug.SetGCPercent(20)
	heap.Monitor()
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
	// Add controller manager.
	k8sManager, aErr := k8s.NewManager()
	if aErr != nil {
		err = aErr
		return
	}
	err = controller.Add(k8sManager, db)
	if err != nil {
		return
	}
	//
	// k8s client.
	client, err := k8s.NewClient()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	// Build Auth
	domain := auth.NewTenant(db, client)
	err = domain.Load()
	if err != nil {
		return
	}
	p, err := auth.New(db, domain)
	if err != nil {
		return
	}
	auth.SetDomain(domain)
	auth.SetIdp(p)
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
	// Build Managers.
	taskManager := task.New(db, client)
	reaperManager := reaper.Manager{
		Client: client,
		DB:     db,
	}
	importManager := importer.Manager{
		DB:          db,
		TaskManager: taskManager,
		Client:      client,
	}
	trackerManager := tracker.Manager{
		DB: db,
	}
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
		metricsManager.Run(ctx)
	}
	// Web
	router := gin.Default()
	router.Use(
		func(ctx *gin.Context) {
			rtx := api.RichContext(ctx)
			rtx.TaskManager = taskManager
			rtx.DB = db
			rtx.Client = client
			defer rtx.Detach()
			ctx.Next()
		})
	router.Use(api.Render())
	router.Use(api.ErrorHandler())
	for _, h := range api.All() {
		h.AddRoutes(router)
	}
	for _, h := range frontend.ALL() {
		h.AddRoutes(router)
	}
	//
	// Auth domain.
	err = domain.Seed()
	if err != nil {
		return
	}
	err = auth.Idp().Cache().Refresh()
	if err != nil {
		return
	}
	// Run Managers.
	k8sManager.Run(ctx)
	taskManager.Run(ctx)
	reaperManager.Run(ctx)
	importManager.Run(ctx)
	trackerManager.Run(ctx)
	// Run Router.
	err = Run(router)
}
