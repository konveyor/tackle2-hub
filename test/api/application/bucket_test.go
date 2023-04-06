package application

import (
	"testing"

	c "github.com/konveyor/tackle2-hub/test/api/client"
)

func TestApplicationBucket(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Samples()[0]

	// Create the application.
	c.Must(t, Create(&application))

	// Bucket test TODO// with Client.BucketGet, BucketPut etc.

	// Clean the application.
	c.Must(t, Delete(&application))
}
