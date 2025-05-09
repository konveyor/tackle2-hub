package manifest

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

var (
	Decrypted = binding.Param{Key: api.Decrypted, Value: "1"}
	Injected  = binding.Param{Key: api.Injected, Value: "1"}
)

func TestManifestCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("", func(t *testing.T) {
			application := &api.Application{Name: t.Name()}
			err := RichClient.Application.Create(application)
			assert.Should(t, err)
			defer func() {
				_ = RichClient.Application.Delete(application.ID)
			}()

			// Create.
			r.Application.ID = application.ID
			err = Manifest.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Manifest.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !assert.MapEq(got.Content, r.Content) || !assert.MapEq(got.Secret, r.Secret) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}
			got, err = Manifest.Get(r.ID, Decrypted)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !assert.MapEq(got.Content, r.Content) {
				t.Errorf("Different content error. Got %v, expected %v", got, r)
			}
			if !assert.MapEq(got.Secret, r.Secret) {
				t.Errorf("Secret not decrypted. Got %v, expected %v", got, r)
			}
			got, err = Manifest.Get(r.ID, Injected)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.MapEq(got.Content, r.Content) {
				t.Errorf("Secret not injected in content. Got %v, expected %v", got, r)
			}

			// Update.
			r.Content["fqdn"] = "hammer"
			err = Manifest.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}
			got, err = Manifest.Get(r.ID, Decrypted)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !assert.MapEq(got.Content, r.Content) || !assert.MapEq(got.Secret, r.Secret) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
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
		})
	}
}
