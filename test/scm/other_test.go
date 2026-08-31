package scm

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
	"github.com/konveyor/tackle2-hub/shared/scm"
	"github.com/onsi/gomega"
)

const (
	EnvHubURL      = "HUB_BASE_URL"
	EnvHubUser     = "BINDING_USER"
	EnvHubPassword = "BINDING_PASSWORD"
)

// newClient builds a hub RichClient from env vars.
// The test is skipped when the hub is not reachable.
func newClient(t *testing.T) (client *binding.RichClient) {
	t.Helper()
	hubURL := os.Getenv(EnvHubURL)
	if hubURL == "" {
		hubURL = "http://localhost:8080"
	}
	user := os.Getenv(EnvHubUser)
	password := os.Getenv(EnvHubPassword)
	if user == "" {
		user = "admin"
	}
	if password == "" {
		password = "admin"
	}
	client = binding.New(hubURL)
	client.Client.SetRetry(1)
	client.Client.Transport().TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client.Client.Use(auth.NewBasic(user, password))
	_, err := client.Setting.List()
	if err != nil {
		t.Skipf("hub not reachable at %s: %v", hubURL, err)
	}
	return
}

// TestInsecure verifies that Insecure reads the kind-specific
// insecure setting from the hub.
func TestInsecure(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	client := newClient(t)

	cases := []struct {
		kind string
		key  string
	}{
		{kind: "git", key: "git.insecure.enabled"},
		{kind: "svn", key: "svn.insecure.enabled"},
	}
	for _, tc := range cases {
		key := tc.key
		repo := api.Repository{Kind: tc.kind}

		original, err := client.Setting.Bool(key)
		g.Expect(err).To(gomega.BeNil())
		t.Cleanup(func() {
			_ = client.Setting.Update(&api.Setting{Key: key, Value: original})
		})

		err = client.Setting.Update(&api.Setting{Key: key, Value: true})
		g.Expect(err).To(gomega.BeNil())
		enabled, err := scm.Insecure(client, repo)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(enabled).To(gomega.BeTrue())

		err = client.Setting.Update(&api.Setting{Key: key, Value: false})
		g.Expect(err).To(gomega.BeNil())
		enabled, err = scm.Insecure(client, repo)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(enabled).To(gomega.BeFalse())
	}
}

// TestProxies verifies that Proxies returns enabled proxies keyed
// by kind with their identity resolved.
func TestProxies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	client := newClient(t)

	identity := &api.Identity{
		Kind:     "proxy",
		Name:     "test-proxy-identity",
		User:     "proxy-user",
		Password: "proxy-password",
	}
	err := client.Identity.Create(identity)
	g.Expect(err).To(gomega.BeNil())
	t.Cleanup(func() {
		_ = client.Identity.Delete(identity.ID)
	})

	proxy, err := client.Proxy.Find("https")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(proxy).NotTo(gomega.BeNil())

	original := *proxy
	t.Cleanup(func() {
		_ = client.Proxy.Update(&original)
	})

	proxy.Enabled = true
	proxy.Host = "proxy.example.com"
	proxy.Port = 8443
	proxy.Excluded = []string{"internal.example.com"}
	proxy.Identity = &api.Ref{ID: identity.ID, Name: identity.Name}
	err = client.Proxy.Update(proxy)
	g.Expect(err).To(gomega.BeNil())

	pm, err := scm.Proxies(client)
	g.Expect(err).To(gomega.BeNil())

	p, found := pm["https"]
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(p.Host).To(gomega.Equal("proxy.example.com"))
	g.Expect(p.Port).To(gomega.Equal(8443))
	g.Expect(p.Excluded).To(gomega.ContainElement("internal.example.com"))
	g.Expect(p.Identity).NotTo(gomega.BeNil())
	g.Expect(p.Identity.ID).To(gomega.Equal(identity.ID))
	g.Expect(p.Identity.User).To(gomega.Equal("proxy-user"))
}

// TestProxiesExcludesDisabled verifies that disabled proxies are omitted.
func TestProxiesExcludesDisabled(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	client := newClient(t)

	proxy, err := client.Proxy.Find("http")
	g.Expect(err).To(gomega.BeNil())
	g.Expect(proxy).NotTo(gomega.BeNil())

	original := *proxy
	t.Cleanup(func() {
		_ = client.Proxy.Update(&original)
	})

	proxy.Enabled = false
	err = client.Proxy.Update(proxy)
	g.Expect(err).To(gomega.BeNil())

	pm, err := scm.Proxies(client)
	g.Expect(err).To(gomega.BeNil())

	_, found := pm["http"]
	g.Expect(found).To(gomega.BeFalse())
}
