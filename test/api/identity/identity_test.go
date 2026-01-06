package identity

import (
	"testing"

	api2 "github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestFindIdentity(t *testing.T) {
	// Setup.
	direct := &api2.Identity{
		Name: "direct",
		Kind: "Test",
	}
	err := Identity.Create(direct)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(direct.ID)
	}()
	direct2 := &api2.Identity{
		Name: "direct2",
		Kind: "Other",
	}
	err = Identity.Create(direct2)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(direct2.ID)
	}()
	application := &api2.Application{
		Name:       t.Name(),
		Identities: []api2.IdentityRef{{ID: direct.ID}},
	}
	err = RichClient.Application.Create(application)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Application.Delete(application.ID)
	}()
	// Find direct.
	filter := binding.Filter{}
	filter.And("application.id").Eq(int(application.ID))
	filter.And("kind").Eq(direct.Kind)
	found, err := Identity.Find(filter)
	assert.Must(t, err)
	if len(found) > 0 {
		identity := found[0]
		if identity.ID != direct.ID {
			t.Errorf("find direct expected: id=%d", direct.ID)
		}
	} else {
		t.Errorf("direct not found")
	}
}
