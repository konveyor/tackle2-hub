package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/api/client"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationBucket(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Samples()[0]

	// Create the application.
	assert.Must(t, Create(&application))

	// Get the bucket to check if it was created.
	err := Client.BucketGet(client.Path(api.BucketRoot, client.Params{api.ID: application.Bucket.ID}), "/dev/null")
	if err != nil {
		t.Errorf(err.Error())
	}

	// Clean the application.
	assert.Must(t, Delete(&application))
}
