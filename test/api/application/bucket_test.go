package application

import (
	"testing"
)

func TestApplicationBucket(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := CloneSamples()[0]

	// Create the application.
	Create(t, application)

	// Bucket test TODO

	// Clean the application.
	Delete(t, application)
}
