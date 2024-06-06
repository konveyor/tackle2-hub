package binding

import (
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/settings"
)

var (
	Settings = &settings.Settings
	Log      = logr.WithName("binding")
)

func init() {
	err := Settings.Load()
	if err != nil {
		panic(err)
	}
}

// The RichClient provides API integration.
type RichClient struct {
	Addon            Addon
	Application      Application
	Archetype        Archetype
	Assessment       Assessment
	Bucket           Bucket
	BusinessService  BusinessService
	Dependency       Dependency
	File             File
	Identity         Identity
	JobFunction      JobFunction
	MigrationWave    MigrationWave
	Proxy            Proxy
	Questionnaire    Questionnaire
	Review           Review
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
	// REST client.
	Client *Client
}

// New builds a new RichClient object.
func New(baseUrl string) (r *RichClient) {
	//
	// Build REST client.
	client := NewClient(baseUrl, api.Login{})

	//
	// Build RichClient.
	r = &RichClient{
		Addon: Addon{
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
		Identity: Identity{
			client: client,
		},
		JobFunction: JobFunction{
			client: client,
		},
		MigrationWave: MigrationWave{
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

	Log.Info("Hub RichClient created.")

	return
}

// Login set token.
func (r *RichClient) Login(user, password string) (err error) {
	login := api.Login{User: user, Password: password}
	err = r.Client.Post(api.AuthLoginRoot, &login)
	if err != nil {
		return
	}
	r.Client.SetToken(login)
	return
}
