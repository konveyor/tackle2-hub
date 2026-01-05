package proxy

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Set of valid resources for tests and reuse.
var (
	PublicHTTP = api.Proxy{
		Kind: "http",
		Host: "http-proxy.local",
		Port: 80,
	}
	PublicHTTPS = api.Proxy{
		Kind:     "https",
		Host:     "https-proxy.local",
		Port:     443,
		Excluded: []string{"excldomain.tld"},
	}
	//PrivateSquid = api.Proxy{
	//	Kind: "https",
	//	Host: "squidprivateproxy.local",
	//	Port: 3128,
	//	Identity: ,
	//}
	Samples = []api.Proxy{PublicHTTP, PublicHTTPS}
)
