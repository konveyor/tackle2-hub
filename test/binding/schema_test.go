package binding

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

const (
	name = "cloudfoundry-coordinates"
)

func TestGetSchema(t *testing.T) {
	g := NewGomegaWithT(t)
	schema := api.RestAPI{}
	err := client.Client.Get("/schema", &schema)
	g.Expect(err).To(BeNil())
	g.Expect(len(schema.Routes) > 0).To(BeTrue())
}

func TestSchemaGet(t *testing.T) {
	g := NewGomegaWithT(t)
	r, err := client.Schema.Get(name)
	g.Expect(err).To(BeNil())
	g.Expect(r.Name).To(Equal(name))
}

func TestSchemaFind(t *testing.T) {
	g := NewGomegaWithT(t)
	r, err := client.Schema.Find("platform", "cloudfoundry", "coordinates")
	g.Expect(err).To(BeNil())
	g.Expect(r.Name).To(Equal(name))
}
