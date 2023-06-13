package identity

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestIdentityCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Identity.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Identity.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = Identity.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Identity.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = Identity.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = Identity.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestIdentityList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Identity.Create(&sample))
		samples[name] = sample
	}

	got, err := Identity.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Identity.Delete(r.ID))
	}
}

func TestIdentityNotCreateDuplicates(t *testing.T) {
	r := GitPw

	// Create sample.
	assert.Should(t, Identity.Create(&r))

	// Try duplicate with the same and different Kind
	for _, kind := range []string{r.Kind, "mvn"} {
		t.Run(kind, func(t *testing.T) {
			// Prepare Identity with duplicate Name.
			dup := &api.Identity{
				Name: r.Name,
				Kind: kind,
			}

			// Try create the duplicate.
			err := Identity.Create(dup)
			if err == nil {
				t.Errorf("Created duplicate identity: %v", dup)

				// Clean the duplicate.
				assert.Must(t, Identity.Delete(dup.ID))
			}
		})
	}

	// Clean.
	assert.Must(t, Identity.Delete(r.ID))
}
