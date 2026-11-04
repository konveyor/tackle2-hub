package jsd

import (
	"testing"

	"github.com/konveyor/tackle2-hub/internal/k8s/simulator"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/onsi/gomega"
)

func TestValidatorFields(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	type Resource struct {
		Name     string
		Pointer  *api.Document
		Value    api.Document
		Nil      *api.Document
		Internal *Document
		hidden   *api.Document
	}
	r := &Resource{
		Pointer: &api.Document{
			Content: api.Map{"name": "A"},
			Schema:  "s1",
		},
		Value: api.Document{
			Content: api.Map{"name": "B"},
			Schema:  "s2",
		},
		Internal: &Document{
			Content: Map{"name": "C"},
			Schema:  "s3",
		},
		hidden: &api.Document{Schema: "s4"},
	}
	v := &Validator{}
	fields := v.fields(r)
	g.Expect(len(fields)).To(gomega.Equal(3))
	g.Expect(fields[0].Schema).To(gomega.Equal("s1"))
	g.Expect(fields[0].Content["name"]).To(gomega.Equal("A"))
	g.Expect(fields[1].Schema).To(gomega.Equal("s2"))
	g.Expect(fields[1].Content["name"]).To(gomega.Equal("B"))
	g.Expect(fields[2].Schema).To(gomega.Equal("s3"))
}

func TestValidatorValidate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Namespace = "konveyor-tackle"
	type Resource struct {
		Coordinates *api.Document
	}
	v := New(simulator.New())
	valid := &Resource{
		Coordinates: &api.Document{
			Schema: "cloudfoundry-coordinates",
			Content: api.Map{
				"organization": "org",
				"space":        "space",
				"name":         "app",
			},
		},
	}
	err := v.Validate(valid)
	g.Expect(err).To(gomega.BeNil())
	invalid := &Resource{
		Coordinates: &api.Document{
			Schema: "cloudfoundry-coordinates",
			Content: api.Map{
				"name": "app",
			},
		},
	}
	err = v.Validate(invalid)
	g.Expect(err).ToNot(gomega.BeNil())
	err = v.Validate(&Resource{})
	g.Expect(err).To(gomega.BeNil())
}
