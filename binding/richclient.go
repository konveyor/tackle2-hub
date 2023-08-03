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
	// Resources APIs.
	Application      Application
	Bucket           Bucket
	BusinessService  BusinessService
	File             File
	Identity         Identity
	JobFunction      JobFunction
	Proxy            Proxy
	RuleSet          RuleSet
	Setting          Setting
	Stakeholder      Stakeholder
	StakeholderGroup StakeholderGroup
	Tag              Tag
	TagCategory      TagCategory
	Target           Target
	Task             Task
	Dependency       Dependency

	// A REST client.
	Client *Client
}

// New builds a new RichClient object.
func New(baseUrl string) (r *RichClient) {
	//
	// Build REST client.
	client := NewClient(baseUrl, "")

	//
	// Build RichClient.
	r = &RichClient{
		Application: Application{
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
		Proxy: Proxy{
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
		Client: client,
	}

	Log.Info("Hub RichClient created.")

	return
}

//
// Login set token.
func (r *RichClient) Login(user, password string) (err error) {
	login := api.Login{User: user, Password: password}
	err = r.Client.Post(api.AuthLoginRoot, &login)
	if err != nil {
		return
	}
	r.Client.SetToken(login.Token)
	return
}
