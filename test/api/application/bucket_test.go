package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationBucket(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Minimal

	// Create the application.
	assert.Must(t, Application.Create(&application))

	// Get the bucket to check if it was created.
	err := Client.BucketGet(binding.Path(api.BucketRoot).Inject(binding.Params{api.ID: application.Bucket.ID}), "/dev/null")
	if err != nil {
		t.Errorf(err.Error())
	}

	// Clean the application.
	assert.Must(t, Application.Delete(application.ID))
}
