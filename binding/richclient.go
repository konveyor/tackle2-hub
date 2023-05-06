package binding

import (
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/settings"
)

var (
	Settings = &settings.Settings
	Log      = logging.WithName("binding")
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
	BusinessService  BusinessService
	Identity         Identity
	JobFunction      JobFunction
	Proxy            Proxy
	Stakeholder      Stakeholder
	StakeholderGroup StakeholderGroup
	Tag              Tag
	TagCategory      TagCategory
	Task             Task

	// A REST client.
	client *Client
}

// Client provides the REST client.
func (h *RichClient) Client() *Client {
	return h.client
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
			client: client,
		},
		BusinessService: BusinessService{
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
		Task: Task{
			client: client,
		},
		client: client,
	}

	Log.Info("Hub RichClient created.")

	return
}

func (r *RichClient) Login(user, password string) (err error) {
	//
	// Build REST client.
	login := api.Login{User: user, Password: password}

	// Login.
	err = r.client.Post(api.AuthLoginRoot, &login)
	if err != nil {
		return
	}
	r.client.SetToken(login.Token)
	return
}
