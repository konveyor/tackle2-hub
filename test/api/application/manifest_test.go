package application

import (
	"encoding/json"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/api/manifest"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestAppManifestGet(t *testing.T) {
	var r api.Manifest
	b, _ := json.Marshal(manifest.Base)
	_ = json.Unmarshal(b, &r)
	// application
	application := &api.Application{Name: t.Name()}
	err := RichClient.Application.Create(application)
	assert.Should(t, err)
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

	if !assert.MapEq(manifest.Base.Content, created.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", manifest.Base.Content, created.Content)
	}
	if assert.MapEq(manifest.Base.Secret, created.Secret) {
		t.Errorf("Secret not encrypted.\n Expected: %s\n Actual: %s", manifest.Base.Secret, created.Secret)
	}

	// Get encrypted.
	got, err := Manifest.Get()
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
	decrypted, err := Manifest.Get(manifest.Decrypted)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.MapEq(created.Content, decrypted.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", created.Content, decrypted.Content)
	}
	if !assert.MapEq(manifest.Base.Secret, decrypted.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", manifest.Base.Secret, decrypted.Secret)
	}
	// Get decrypted and injected.
	injected, err := Manifest.Get(manifest.Decrypted, manifest.Injected)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.MapEq(manifest.InjectedContent, injected.Content) {
		t.Errorf("Content not injected.\n Expected: %s\n Actual: %s", manifest.InjectedContent, injected.Content)
	}
	if assert.MapEq(created.Secret, injected.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", created.Secret, injected.Secret)
	}
}

func TestAppManifestGet_Select(t *testing.T) {
	var r api.Manifest
	b, _ := json.Marshal(manifest.Base)
	_ = json.Unmarshal(b, &r)
	// application
	application := &api.Application{Name: t.Name()}
	err := RichClient.Application.Create(application)
	assert.Should(t, err)
	defer func() {
		_ = Application.Delete(application.ID)
	}()

	// Create.
	r.Application.ID = application.ID
	err = Application.Select(application.ID).Manifest.Create(&r)
	if err != nil {
		t.Errorf(err.Error())
	}
	defer func() {
		_ = RichClient.Manifest.Delete(r.ID)
	}()
	created := r

	if !assert.MapEq(manifest.Base.Content, created.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", manifest.Base.Content, created.Content)
	}
	if assert.MapEq(manifest.Base.Secret, created.Secret) {
		t.Errorf("Secret not encrypted.\n Expected: %s\n Actual: %s", manifest.Base.Secret, created.Secret)
	}

	// Get encrypted.
	got, err := Application.Select(application.ID).Manifest.Get()
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
	decrypted, err := Application.Select(application.ID).Manifest.Get(manifest.Decrypted)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.MapEq(created.Content, decrypted.Content) {
		t.Errorf("Content mismatch.\n Expected: %s\n Actual: %s", created.Content, decrypted.Content)
	}
	if !assert.MapEq(manifest.Base.Secret, decrypted.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", manifest.Base.Secret, decrypted.Secret)
	}
	// Get decrypted and injected.
	injected, err := Application.Select(application.ID).Manifest.Get(manifest.Decrypted, manifest.Injected)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.MapEq(manifest.InjectedContent, injected.Content) {
		t.Errorf("Content not injected.\n Expected: %s\n Actual: %s", manifest.InjectedContent, injected.Content)
	}
	if assert.MapEq(created.Secret, injected.Secret) {
		t.Errorf("Secret not decrypted.\n Expected: %s\n Actual: %s", created.Secret, injected.Secret)
	}
}
