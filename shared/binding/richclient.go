package binding

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/shared/api"
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
	AnalysisProfile  AnalysisProfile
	Application      Application
	Archetype        Archetype
	Assessment       Assessment
	Bucket           Bucket
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
	Task             Task
	Ticket           Ticket
	Tracker          Tracker
}

// New builds a new RichClient object.
func New(baseURL string) (r *RichClient) {
	client := NewClient(baseURL)
	r = &RichClient{
		Addon: Addon{
			client: client,
		},
		AnalysisProfile: AnalysisProfile{
			client: client,
		},
		Application: Application{
			client: client,
		},
		Archetype: Archetype{
			client: client,
		},
		Assessment: Assessment{
			client: client,
		},
		Bucket: Bucket{
			client: client,
		},
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
		Task: Task{
			client: client,
		},
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
