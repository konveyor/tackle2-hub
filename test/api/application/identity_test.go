package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
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
	indirect := &api.Identity{
		Kind:    "Other",
		Name:    "indirect",
		Default: true,
	}
	err = RichClient.Identity.Create(indirect)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(indirect.ID)
	}()
	indirect2 := &api.Identity{
		Kind:    "Test",
		Name:    "indirect-shadowed",
		Default: true,
	}
	err = RichClient.Identity.Create(indirect2)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(indirect2.ID)
	}()
	application := &api.Application{
		Name:       t.Name(),
		Identities: []api.IdentityRef{{ID: direct.ID, Role: "asset"}},
	}
	err = Application.Create(application)
	assert.Must(t, err)
	defer func() {
		_ = Application.Delete(application.ID)
	}()
	// Find direct.
	filter := binding.Filter{}
	filter.And("kind").Eq(direct.Kind)
	identity, found, err := Application.Identity(application.ID).Find(filter)
	assert.Must(t, err)
	if found {
		if identity.ID != direct.ID {
			t.Errorf("find direct expected: id=%d", direct.ID)
		}
	} else {
		t.Errorf("direct not found")
	}
	// Find indirect.
	filter = binding.Filter{}
	filter.And("kind").Eq(indirect.Kind)
	identity, found, err = Application.Identity(application.ID).Find(filter)
	assert.Must(t, err)
	if found {
		if identity.ID != indirect.ID {
			t.Errorf("find indirect expected: id=%d", indirect.ID)
		}
	} else {
		t.Errorf("indirect not found")
	}
	// Not find indirect.
	filter = binding.Filter{}
	filter.And("kind").Eq("none")
	_, found, err = Application.Identity(application.ID).Find(filter)
	assert.Must(t, err)
	if found {
		t.Errorf("not found expected")
	}
	// List
	list, err := Application.Identity(application.ID).List()
	assert.Must(t, err)
	if len(list) != 2 {
		t.Errorf("list expected: 1, actual: %d", len(list))
	}
}
