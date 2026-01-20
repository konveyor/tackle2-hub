package binding

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
	"github.com/onsi/gomega"
)

func TestStub(t *testing.T) {
	g := gomega.NewWithT(t)
	binding := New("")
	binding.Use(&client.Stub{
		GetFn: func(path string, object any, params ...Param) (err error) {
			switch r := object.(type) {
			case *api.Application:
				r.ID = 23
				r.Name = "Test"
			default:
				err = &NotFound{}
			}
			return
		},
	})

	application, err := binding.Application.Get(1)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(application.ID).To(gomega.Equal(uint(23)))
	g.Expect(application.Name).To(gomega.Equal("Test"))
}
