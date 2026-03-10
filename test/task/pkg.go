package task

import (
	"context"
	"time"

	"github.com/konveyor/tackle2-hub/internal/k8s/simulator"
	"github.com/konveyor/tackle2-hub/internal/model"
	impTask "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Context is the test context.
type Context struct {
	DB       *gorm.DB
	Client   client.Client
	Manager  *impTask.Manager
	Cancel   context.CancelFunc
	Captured struct {
		TaskFrequency time.Duration
	}
}

// setup creates and configures a complete test environment.
func setup(g *gomega.GomegaWithT) (ctx *Context) {
	ctx = &Context{}

	// Adjust settings for faster test execution
	ctx.Captured.TaskFrequency = settings.Settings.Frequency.Task
	settings.Settings.Frequency.Task = 100 * time.Millisecond

	// Setup in-memory database
	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	g.Expect(err).To(gomega.BeNil())
	ctx.DB = db

	// Auto-migrate required tables
	err = db.AutoMigrate(
		&model.Task{},
		&model.TaskGroup{},
		&model.Application{},
		&model.Bucket{},
		&model.File{},
		&model.TagCategory{},
		&model.Tag{},
	)
	g.Expect(err).To(gomega.BeNil())

	// Create k8s simulator with fast timing for testing
	// Pods will transition: Pending (1s) -> Running (1s) -> Succeeded
	ctx.Client = simulator.New().Use(simulator.NewManager(1, 1))

	return
}

// teardown cleans up the test environment.
func (ctx *Context) teardown() {
	// Restore original settings
	settings.Settings.Frequency.Task = ctx.Captured.TaskFrequency

	// Cancel context to stop manager if running
	if ctx.Cancel != nil {
		ctx.Cancel()
		// Give manager time to stop gracefully
		time.Sleep(50 * time.Millisecond)
	}
}

// seed seeds the database with common test data.
func (ctx *Context) seed(g *gomega.GomegaWithT) (app *model.Application) {
	// Create TagCategory and Tag
	category := &model.TagCategory{
		Name: "Language",
	}
	err := ctx.DB.Create(category).Error
	g.Expect(err).To(gomega.BeNil())
	tag := &model.Tag{
		Name:     "Java",
		Category: *category,
	}
	err = ctx.DB.Create(tag).Error
	g.Expect(err).To(gomega.BeNil())

	// Create an application tagged with Java
	app = &model.Application{
		Name: "Test Application",
		Tags: []model.Tag{*tag},
	}
	err = ctx.DB.Create(app).Error
	g.Expect(err).To(gomega.BeNil())

	return
}

// newManager creates and starts the task manager.
func (ctx *Context) newManager(g *gomega.GomegaWithT) {
	ctx.Manager = &impTask.Manager{
		DB:     ctx.DB,
		Client: ctx.Client,
		Scopes: []string{},
	}
	managerCtx, cancel := context.WithCancel(context.Background())
	ctx.Cancel = cancel

	// Run the manager in a goroutine
	go ctx.Manager.Run(managerCtx)

	// Wait for the cluster to refresh and be ready
	// With Settings.Frequency.Task at 100ms, this gives time for initial setup
	time.Sleep(300 * time.Millisecond)
}
