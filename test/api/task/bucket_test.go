package task

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTaskBucket(t *testing.T) {
	task := Windup
	// Create the application.
	assert.Must(t, Task.Create(&task))
	// Get the bucket to check if it was created.
	destDir, err := ioutil.TempDir("", "destDir")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer func() {
		_ = os.RemoveAll(destDir)
	}()
	bucket := RichClient.Task.Bucket(task.ID)
	err = bucket.Get("", destDir)
	if err != nil {
		t.Errorf(err.Error())
	}
	// Clean the application.
	assert.Must(t, Task.Delete(task.ID))
}
