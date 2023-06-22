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
	Stakeholder      Stakeholder
	StakeholderGroup StakeholderGroup
	Tag              Tag
	TagCategory      TagCategory
	Task             Task
	Dependency       Dependency

	// A REST client.
	Client *Client
}

// newRichClient builds a new RichClient object.
func New(baseUrl string) (r *RichClient) {
	//
	// Build REST client.
	client := NewClient(baseUrl, "")

	//
	// Build RichClient.
	r = &RichClient{
		Application: Application{
			Client: client,
		},
		Bucket: Bucket{
			Client: client,
		},
		BusinessService: BusinessService{
			Client: client,
		},
		Dependency: Dependency{
			Client: client,
		},
		File: File{
			Client: client,
		},
		Identity: Identity{
			Client: client,
		},
		JobFunction: JobFunction{
			Client: client,
		},
		Proxy: Proxy{
			Client: client,
		},
		RuleSet: RuleSet{
			Client: client,
		},
		Stakeholder: Stakeholder{
			Client: client,
		},
		StakeholderGroup: StakeholderGroup{
			Client: client,
		},
		Tag: Tag{
			Client: client,
		},
		TagCategory: TagCategory{
			Client: client,
		},
		Task: Task{
			Client: client,
		},
		Client: client,
	}

	Log.Info("Hub RichClient created.")

	return
}

func (r *RichClient) Login(user, password string) (err error) {
	//
	// Build REST client.
	login := api.Login{User: user, Password: password}

	// Login.
	err = r.Client.Post(api.AuthLoginRoot, &login)
	if err != nil {
		return
	}
	r.Client.SetToken(login.Token)
	return
}
