package binding

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAddon(t *testing.T) {
	g := NewGomegaWithT(t)

	// LIST: List all addons
	list, err := client.Addon.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(BeNumerically(">=", 3)) // At least 3 seeded addons

	// Verify analyzer addon exists
	foundAnalyzer := false
	for _, addon := range list {
		if addon.Name == "analyzer" {
			foundAnalyzer = true
			g.Expect(addon.Container).NotTo(BeNil())
			g.Expect(addon.Container.Image).To(ContainSubstring("tackle2-addon-analyzer"))
			break
		}
	}
	g.Expect(foundAnalyzer).To(BeTrue())

	// GET: Retrieve specific addon by name
	analyzer, err := client.Addon.Get("analyzer")
	g.Expect(err).To(BeNil())
	g.Expect(analyzer).NotTo(BeNil())
	g.Expect(analyzer.Name).To(Equal("analyzer"))
	g.Expect(analyzer.Container).NotTo(BeNil())
	g.Expect(analyzer.Container.Image).To(ContainSubstring("tackle2-addon-analyzer"))
	g.Expect(analyzer.Container.Name).To(Equal("addon"))

	// GET: Retrieve language-discovery addon
	languageDiscovery, err := client.Addon.Get("language-discovery")
	g.Expect(err).To(BeNil())
	g.Expect(languageDiscovery).NotTo(BeNil())
	g.Expect(languageDiscovery.Name).To(Equal("language-discovery"))
	g.Expect(languageDiscovery.Container).NotTo(BeNil())
	g.Expect(languageDiscovery.Container.Image).To(ContainSubstring("tackle2-addon-discovery"))

	// GET: Retrieve platform addon
	platform, err := client.Addon.Get("platform")
	g.Expect(err).To(BeNil())
	g.Expect(platform).NotTo(BeNil())
	g.Expect(platform.Name).To(Equal("platform"))
	g.Expect(platform.Container).NotTo(BeNil())
	g.Expect(platform.Container.Image).To(ContainSubstring("tackle2-addon-platform"))

	// GET: Try to retrieve non-existent addon
	_, err = client.Addon.Get("non-existent-addon")
	g.Expect(err).NotTo(BeNil())
}
