package main

import (
	"context"
	"encoding/json"
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
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/reaper"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
	"github.com/konveyor/tackle2-hub/volume"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	"io/ioutil"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
	"os"
	"path"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"
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
	syncHelper := model.SyncHelper{}
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
	var provider auth.Provider
	if settings.Settings.Auth.Required {
		k := auth.NewKeycloak(
			settings.Settings.Auth.Keycloak.Host,
			settings.Settings.Auth.Keycloak.Realm,
			settings.Settings.Auth.Keycloak.ClientID,
			settings.Settings.Auth.Keycloak.ClientSecret)
		provider = &k
	} else {
		provider = &auth.NoAuth{}
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
	importManager.Run(context.Background(), &syncHelper.DependencyMutex)
	//
	// Web
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	for _, h := range api.All() {
		h.With(db, client, provider, &syncHelper)
		h.AddRoutes(router)
	}
	err = router.Run()
}

//
// Seed the database with the contents of json
// files contained in DB_SEED_PATH.
func seed(db *gorm.DB, models []interface{}) (err error) {
	for _, m := range models {
		err = func() (err error) {
			kind := reflect.TypeOf(m).Name()
			fileName := strings.ToLower(kind) + ".json"
			filePath := path.Join(settings.Settings.DB.SeedPath, fileName)
			file, err := os.Open(filePath)
			if err != nil {
				err = nil
				return
			}
			defer file.Close()
			jsonBytes, err := ioutil.ReadAll(file)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}

			var unmarshalled []map[string]interface{}
			err = json.Unmarshal(jsonBytes, &unmarshalled)
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
			for i := range unmarshalled {
				result := db.Model(&m).Create(unmarshalled[i])
				if result.Error != nil {
					err = result.Error
					return
				}
			}
			return
		}()
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}

	seeded, _ := json.Marshal(true)
	setting := model.Setting{Key: ".hub.db.seeded", Value: seeded}
	result := db.Create(&setting)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	log.Info("Database seeded.")
	return
}
