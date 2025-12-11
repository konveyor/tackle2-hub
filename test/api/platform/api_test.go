package manifest

import (
	"encoding/json"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestPlatformCRUD(t *testing.T) {
	var r api.Platform
	b, _ := json.Marshal(Base)
	_ = json.Unmarshal(b, &r)
	// application
	application := &api.Application{Name: t.Name()}
	err := RichClient.Application.Create(application)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Application.Delete(application.ID)
	}()
	// identity
	identity := &api.Identity{
		Name: t.Name(),
		Kind: t.Name(),
	}
	err = RichClient.Identity.Create(identity)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(identity.ID)
	}()

	// Create.
	r.Identity = &api.Ref{
		ID:   identity.ID,
		Name: identity.Name,
	}
	err = Platform.Create(&r)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Get
	got, err := Platform.Get(r.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.Eq(got, r) {
		t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
	}

	// Update.
	r.Name = r.Name + "updated"
	err = Platform.Update(&r)
	if err != nil {
		t.Errorf(err.Error())
	}
	got, err = Platform.Get(r.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.UpdateUser = got.UpdateUser
	if !assert.Eq(got, r) {
		t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
	}

	// Delete.
	err = Platform.Delete(r.ID)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = Platform.Get(r.ID)
	if err == nil {
		t.Errorf("Resource exits, but should be deleted: %v", r)
	}
}
