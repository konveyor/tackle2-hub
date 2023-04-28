package binding

import (
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/settings"
)

var (
	Settings = &settings.Settings
	Log      = logging.WithName("binding")
)

var (
	// Addon An addon richClient configured for a task execution.
	RichClient *Adapter
	// Client
	//client *Client
)

func init() {
	err := Settings.Load()
	if err != nil {
		panic(err)
	}

	RichClient = newRichClient()
}

// The RichClient provides API integration.
type Adapter struct {
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
func (h *Adapter) Client() *Client {
	return h.client
}

// newRichClient builds a new Addon RichClient object.
func newRichClient() (richClient *Adapter) {
	//
	// Build REST client.
	client := NewClient(Settings.Addon.Hub.URL, Settings.Addon.Hub.Token)
	//
	// Build RichClient.
	richClient = &Adapter{
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

