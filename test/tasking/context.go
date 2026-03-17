package tasking

import (
	"context"

	"github.com/konveyor/tackle2-hub/internal/k8s/simulator"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/task"
	"github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(g *gomega.GomegaWithT) (ctx *Context) {
	ctx = &Context{}
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
		&model.Platform{},
		&model.Bucket{},
		&model.File{},
		&model.TagCategory{},
		&model.Tag{},
	)
	g.Expect(err).To(gomega.BeNil())
	ctx.Client = simulator.New().Use(simulator.NewManager(0, 0))
	ctx.seed(g)
	return
}

// Context is the test context.
type Context struct {
	DB          *gorm.DB
	Client      client.Client
	Manager     *task.Manager
	Application *model.Application
	Platform    *model.Platform
}

// seed seeds the database with common test data.
func (ctx *Context) seed(g *gomega.GomegaWithT) {
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
	ctx.Application = &model.Application{
		Name: "Test Application",
		Tags: []model.Tag{*tag},
	}
	err = ctx.DB.Create(ctx.Application).Error
	g.Expect(err).To(gomega.BeNil())

	// Create a platform for platform-based task tests
	ctx.Platform = &model.Platform{
		Name: "Test Platform",
		Kind: "kubernetes",
	}
	err = ctx.DB.Create(ctx.Platform).Error
	g.Expect(err).To(gomega.BeNil())
	return
}

func (ctx *Context) reconcile(g *gomega.GomegaWithT, n int, taskIds ...uint) {
	maxCycles := max(100, n*10)
	if len(taskIds) == 0 {
		return
	}
	var tasks []*model.Task
	for i := 0; i < maxCycles; i++ {
		_ = ctx.Manager.Reconcile(context.Background())
		err := ctx.DB.Find(&tasks, taskIds).Error
		g.Expect(err).To(gomega.BeNil())

		count := 0
		for _, m := range tasks {
			if m.State == task.Succeeded ||
				m.State == task.Failed ||
				m.State == task.Canceled {
				count++
			}
		}
		if count == n {
			return
		}
	}
	g.Fail("Not ALL tasks completed.")
}
