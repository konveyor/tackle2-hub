package application

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationBucket(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Minimal

	// Create the application.
	assert.Must(t, Application.Create(&application))

	// Get the bucket to check if it was created.
	destDir, err := ioutil.TempDir("", "destDir")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer func() {
		_ = os.RemoveAll(destDir)
	}()
	bucket := RichClient.Application.Bucket(application.ID)
	err = bucket.Get("", destDir)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Clean the application.
	assert.Must(t, Application.Delete(application.ID))
}

func TestApplicationBucket_Select(t *testing.T) {
	// Test Facts subresource on the first sample application only.
	application := Minimal

	// Create the application.
	assert.Must(t, Application.Create(&application))

	// Get the bucket to check if it was created.
	destDir, err := ioutil.TempDir("", "destDir")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer func() {
		_ = os.RemoveAll(destDir)
	}()
	err = RichClient.Application.Select(application.ID).Bucket.Get("", destDir)
	if err != nil {
		t.Errorf(err.Error())
	}

	// Clean the application.
	assert.Must(t, Application.Delete(application.ID))
}
