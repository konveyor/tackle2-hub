package application

import (
	"encoding/json"
	"testing"

	manifest2 "github.com/konveyor/tackle2-hub/internal/test/api/manifest"
	assert2 "github.com/konveyor/tackle2-hub/internal/test/assert"
	"github.com/konveyor/tackle2-hub/shared/api"
)

func TestAppManifestGet(t *testing.T) {
	var r api.Manifest
	b, _ := json.Marshal(manifest2.Base)
	_ = json.Unmarshal(b, &r)
	// application
	application := &api.Application{Name: t.Name()}
	err := RichClient.Application.Create(application)
	assert2.Should(t, err)
	defer func() {
		_ = Application.Delete(application.ID)
	}()

	Manifest := Application.Manifest(application.ID)

	// Create.
	r.Application.ID = application.ID
	err = Manifest.Create(&r)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer func() {
		_ = RichClient.Manifest.Delete(r.ID)
	}()
	created := r

	if !assert2.MapEq(manifest2.Base.Content, created.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", manifest2.Base.Content, created.Content)
	}
	if assert2.MapEq(manifest2.Base.Secret, created.Secret) {
		t.Errorf("Secret not encrypted.\n Expected: %s\n Actual: %s", manifest2.Base.Secret, created.Secret)
	}

	// Get encrypted.
	got, err := Manifest.Get()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert2.MapEq(created.Content, got.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", created.Content, got.Content)
	}
	if !assert2.MapEq(created.Secret, got.Secret) {
		t.Errorf("Secret not encrypted.\n Expected: %s\n Actual: %s", created.Secret, got.Secret)
	}
	// Get decrypted.
	decrypted, err := Manifest.Get(manifest2.Decrypted)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert2.MapEq(created.Content, decrypted.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", created.Content, decrypted.Content)
	}
	if !assert2.MapEq(manifest2.Base.Secret, decrypted.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", manifest2.Base.Secret, decrypted.Secret)
	}
	// Get decrypted and injected.
	injected, err := Manifest.Get(manifest2.Decrypted, manifest2.Injected)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert2.MapEq(manifest2.InjectedContent, injected.Content) {
		t.Errorf("Content not injected.\n Expected: %s\n Actual: %s", manifest2.InjectedContent, injected.Content)
	}
	if assert2.MapEq(created.Secret, injected.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", created.Secret, injected.Secret)
	}
}
