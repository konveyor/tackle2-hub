package binding

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/analysis"
	"github.com/konveyor/tackle2-hub/shared/binding/application"
	"github.com/konveyor/tackle2-hub/shared/binding/archetype"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
	_import "github.com/konveyor/tackle2-hub/shared/binding/import"
	"github.com/konveyor/tackle2-hub/shared/binding/report"
	"github.com/konveyor/tackle2-hub/shared/binding/tagcategory"
	"github.com/konveyor/tackle2-hub/shared/binding/task"
	"github.com/konveyor/tackle2-hub/shared/binding/taskgroup"
)

var (
	Log = logr.New("binding", 0)
)

// New builds a new RichClient object.
func New(baseURL string) (r *RichClient) {
	r = &RichClient{}
	r.build(client.New(baseURL))
	return
}

// The RichClient provides API integration.
type RichClient struct {
	// Client
	Client RestClient
	// API namespaces.
	Addon            Addon
	Analysis         analysis.Analysis
	AnalysisProfile  AnalysisProfile
	Application      application.Application
	Archetype        archetype.Archetype
	Assessment       Assessment
	Bucket           bucket.Bucket
	BusinessService  BusinessService
	ConfigMap        ConfigMap
	Dependency       Dependency
	File             File
	Generator        Generator
	Identity         Identity
	Import           _import.Import
	JobFunction      JobFunction
	Manifest         Manifest
	MigrationWave    MigrationWave
	Platform         Platform
	Proxy            Proxy
	Questionnaire    Questionnaire
	Report           report.Report
	Review           Review
	Schema           Schema
	RuleSet          RuleSet
	Setting          Setting
	Stakeholder      Stakeholder
	StakeholderGroup StakeholderGroup
	Tag              Tag
	TagCategory      tagcategory.TagCategory
	Target           Target
	Task             task.Task
	TaskGroup        taskgroup.TaskGroup
	Ticket           Ticket
	Tracker          Tracker
}

// Use login.
func (r *RichClient) Use(client RestClient) {
	r.build(client)
}

// Login set token.
func (r *RichClient) Login(user, password string) (err error) {
	login := api.Login{
		User:     user,
		Password: password,
	}
	err = r.Client.Post(api.AuthLoginRoute, &login)
	if err != nil {
		return
	}
	r.Client.Use(login)
	return
}

// build the handlers.
func (r *RichClient) build(client RestClient) {
	r.Client = client
	r.Addon = Addon{client: client}
	r.Analysis = analysis.New(client)
	r.AnalysisProfile = AnalysisProfile{client: client}
	r.Application = application.New(client)
	r.Archetype = archetype.New(client)
	r.Assessment = Assessment{client: client}
	r.Bucket = bucket.New(client)
	r.BusinessService = BusinessService{client: client}
	r.ConfigMap = ConfigMap{client: client}
	r.Dependency = Dependency{client: client}
	r.File = File{client: client}
	r.Generator = Generator{client: client}
	r.Identity = Identity{client: client}
	r.Import = _import.New(client)
	r.JobFunction = JobFunction{client: client}
	r.Manifest = Manifest{client: client}
	r.MigrationWave = MigrationWave{client: client}
	r.Platform = Platform{client: client}
	r.Proxy = Proxy{client: client}
	r.Questionnaire = Questionnaire{client: client}
	r.Report = report.New(client)
	r.Review = Review{client: client}
	r.RuleSet = RuleSet{client: client}
	r.Schema = Schema{client: client}
	r.Setting = Setting{client: client}
	r.Stakeholder = Stakeholder{client: client}
	r.StakeholderGroup = StakeholderGroup{client: client}
	r.Tag = Tag{client: client}
	r.TagCategory = tagcategory.New(client)
	r.Target = Target{client: client}
	r.Task = task.New(client)
	r.TaskGroup = taskgroup.New(client)
	r.Ticket = Ticket{client: client}
	r.Tracker = Tracker{client: client}
}
