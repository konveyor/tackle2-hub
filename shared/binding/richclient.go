package binding

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/analysis"
	"github.com/konveyor/tackle2-hub/shared/binding/application"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
	"github.com/konveyor/tackle2-hub/shared/binding/task"
)

var (
	Log = logr.New("binding", 0)
)

// The RichClient provides API integration.
type RichClient struct {
	// Client
	Client *Client
	// API namespaces.
	Addon            Addon
	Analysis         analysis.Analysis
	AnalysisProfile  AnalysisProfile
	Application      application.Application
	Archetype        Archetype
	Assessment       Assessment
	Bucket           bucket.Bucket
	BusinessService  BusinessService
	Dependency       Dependency
	File             File
	Generator        Generator
	Identity         Identity
	JobFunction      JobFunction
	Manifest         Manifest
	MigrationWave    MigrationWave
	Platform         Platform
	Proxy            Proxy
	Questionnaire    Questionnaire
	Review           Review
	Schema           Schema
	RuleSet          RuleSet
	Setting          Setting
	Stakeholder      Stakeholder
	StakeholderGroup StakeholderGroup
	Tag              Tag
	TagCategory      TagCategory
	Target           Target
	Task             task.Task
	Ticket           Ticket
	Tracker          Tracker
}

// New builds a new RichClient object.
func New(baseURL string) (r *RichClient) {
	client := client.New(baseURL)
	r = &RichClient{
		Addon: Addon{
			client: client,
		},
		Analysis: analysis.New(client),
		AnalysisProfile: AnalysisProfile{
			client: client,
		},
		Application: application.New(client),
		Archetype: Archetype{
			client: client,
		},
		Assessment: Assessment{
			client: client,
		},
		Bucket: bucket.New(client),
		BusinessService: BusinessService{
			client: client,
		},
		Dependency: Dependency{
			client: client,
		},
		File: File{
			client: client,
		},
		Generator: Generator{
			client: client,
		},
		Identity: Identity{
			client: client,
		},
		JobFunction: JobFunction{
			client: client,
		},
		Manifest: Manifest{
			client: client,
		},
		MigrationWave: MigrationWave{
			client: client,
		},
		Platform: Platform{
			client: client,
		},
		Proxy: Proxy{
			client: client,
		},
		Questionnaire: Questionnaire{
			client: client,
		},
		Review: Review{
			client: client,
		},
		RuleSet: RuleSet{
			client: client,
		},
		Schema: Schema{
			client: client,
		},
		Setting: Setting{
			client: client,
		},
		Stakeholder: Stakeholder{
			client: client,
		},
		StakeholderGroup: StakeholderGroup{
			client: client,
		},
		Tag: Tag{
			client: client,
		},
		TagCategory: TagCategory{
			client: client,
		},
		Target: Target{
			client: client,
		},
		Task: task.New(client),
		Ticket: Ticket{
			client: client,
		},
		Tracker: Tracker{
			client: client,
		},
		Client: client,
	}
	return
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
	r.Client.Login = login
	return
}
