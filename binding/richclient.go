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
	//	// Task API.
	//	Task
	//	// Settings API
	//	Setting Setting
	// Application API.
	Application Application
	// Identity API.
	//	Identity Identity
	//	// Proxy API.
	//	Proxy Proxy
	//	// TagCategory API.
	//	TagCategory TagCategory
	//	// Tag API.
	//	Tag Tag
	//	// File API.
	//	File File
	//	// RuleBundle API
	//	RuleBundle RuleBundle
	// client A REST client.
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
		//		Task: Task{
		//			client: client,
		//		},
		//		Setting: Setting{
		//			client: client,
		//		},
		Application: Application{
			client: client,
		},
		//		Identity: Identity{
		//			client: client,
		//		},
		//		Proxy: Proxy{
		//			client: client,
		//		},
		//		TagCategory: TagCategory{
		//			client: client,
		//		},
		//		Tag: Tag{
		//			client: client,
		//		},
		//		File: File{
		//			client: client,
		//		},
		//		RuleBundle: RuleBundle{
		//			client: client,
		//		},
		client: client,
	}

	Log.Info("Hub RichClient (adapter) created.")

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