package manifest

import (
	"encoding/json"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestManifestCRUD(t *testing.T) {
	var r api.Manifest
	b, _ := json.Marshal(Base)
	_ = json.Unmarshal(b, &r)
	// application
	application := &api.Application{Name: t.Name()}
	err := RichClient.Application.Create(application)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Application.Delete(application.ID)
	}()

	// Create.
	r.Application.ID = application.ID
	err = Manifest.Create(&r)
	if err != nil {
		t.Fatalf(err.Error())
	}
	created := r
	if !assert.MapEq(Base.Content, created.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", Base.Content, created.Content)
	}
	if assert.MapEq(Base.Secret, created.Secret) {
		t.Errorf("Secret not encrypted.\n Expected: %s\n Actual: %s", Base.Secret, created.Secret)
	}

	// Get encrypted.
	got, err := Manifest.Get(created.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.MapEq(created.Content, got.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", created.Content, got.Content)
	}
	if !assert.MapEq(created.Secret, got.Secret) {
		t.Errorf("Secret not encrypted.\n Expected: %s\n Actual: %s", created.Secret, got.Secret)
	}
	// Get decrypted.
	decrypted, err := Manifest.Get(created.ID, Decrypted)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.MapEq(created.Content, decrypted.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", created.Content, decrypted.Content)
	}
	if !assert.MapEq(Base.Secret, decrypted.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", Base.Secret, decrypted.Secret)
	}
	// Get decrypted and injected.
	injected, err := Manifest.Get(created.ID, Decrypted, Injected)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.MapEq(InjectedContent, injected.Content) {
		t.Errorf("Content not injected.\n Expected: %s\n Actual: %s", InjectedContent, injected.Content)
	}
	if assert.MapEq(created.Secret, injected.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", created.Secret, injected.Secret)
	}

	// Update.
	r.Content["fqdn"] = "hunter.com"
	r.Secret["password"] = "_$44rabbit-"
	err = Manifest.Update(&r)
	if err != nil {
		t.Errorf(err.Error())
	}
	got, err = Manifest.Get(r.ID, Decrypted)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if got.Content["fqdn"] != r.Content["fqdn"] {
		t.Errorf("fqdn not updated.")
	}
	if got.Secret["password"] != r.Secret["password"] {
		t.Errorf("password not updated.")
	}

	// Delete.
	err = Manifest.Delete(r.ID)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = Manifest.Get(r.ID)
	if err == nil {
		t.Errorf("Resource exits, but should be deleted: %v", r)
	}
}
