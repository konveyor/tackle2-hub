package application

import (
	"testing"
)

func TestApplicationBucket(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Samples()[0]

	// Create the application.
	Create(t, application)

	// Clean the application.
	EnsureDelete(t, application)
}
