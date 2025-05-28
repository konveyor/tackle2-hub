package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestFindIdentity(t *testing.T) {
	// Setup.
	direct := &api.Identity{
		Name: "direct",
		Kind: "Test",
	}
	err := RichClient.Identity.Create(direct)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(direct.ID)
	}()
	direct2 := &api.Identity{
		Name: "direct2",
		Kind: "Other",
	}
	err = RichClient.Identity.Create(direct2)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(direct2.ID)
	}()
	inherited := &api.Identity{
		Kind:    "Other",
		Name:    "inherited",
		Default: true,
	}
	err = RichClient.Identity.Create(inherited)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(inherited.ID)
	}()
	application := &api.Application{
		Name:       t.Name(),
		Identities: []api.Ref{{ID: direct.ID}},
	}
	err = RichClient.Application.Create(application)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Application.Delete(application.ID)
	}()
	// Find direct.
	identity, found, err := Application.FindIdentity(application.ID, direct.Kind)
	assert.Must(t, err)
	if !found {
		t.Errorf("not found")
	}
	if identity.ID != direct.ID {
		t.Errorf("find direct expected: id=%d", direct.ID)
	}
	// Find inherited.
	identity, found, err = Application.FindIdentity(application.ID, inherited.Kind)
	assert.Must(t, err)
	if !found {
		t.Errorf("not found")
	}
	if identity.ID != inherited.ID {
		t.Errorf("find inherited expected: id=%d", inherited.ID)
	}
	// Not find inherited.
	_, found, err = Application.FindIdentity(application.ID, "None")
	assert.Must(t, err)
	if found {
		t.Errorf("not found expected")
	}
}
