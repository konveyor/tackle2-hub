package binding

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)

func TestProxy(t *testing.T) {
	g := NewGomegaWithT(t)

	// Define the proxy to create
	proxy := &api.Proxy{
		Resource: api.Resource{ID: 1},
		Kind:     "http",
		Host:     "http-proxy.local",
		Port:     80,
	}

	// GET: Retrieve the proxy and verify it matches
	retrieved, err := client.Proxy.Get(proxy.ID)
	g.Expect(err).To(BeNil())
	g.Expect(retrieved).NotTo(BeNil())

	// UPDATE: Modify the proxy
	proxy.Host = "http-proxy-updated.local"
	proxy.Port = 8080
	proxy.Excluded = []string{"example.com", "localhost"}

	err = client.Proxy.Update(proxy)
	g.Expect(err).To(BeNil())

	// GET: Retrieve again and verify updates
	updated, err := client.Proxy.Get(proxy.ID)
	g.Expect(err).To(BeNil())
	g.Expect(updated).NotTo(BeNil())
	eq, report := cmp.Eq(proxy, updated, "CreateTime", "UpdateUser")
	g.Expect(eq).To(BeTrue(), report)
}
