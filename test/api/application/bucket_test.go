package application

import (
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationBucket(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Samples()[0]

	// Create the application.
	assert.Must(t, Create(&application))

	// Bucket test TODO// with Client.BucketGet, BucketPut etc.

	// Clean the application.
	assert.Must(t, Delete(&application))
}
