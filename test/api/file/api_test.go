package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestFilePutGetDelete(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			origPath, _ := filepath.Abs(r.Path)
			// Create.
			file, err := File.Put(origPath)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			tempFilePath := fmt.Sprintf("/tmp/gotFile-%s", r.Name)
			err = File.Get(file.ID, tempFilePath)
			if err != nil {
				t.Errorf(err.Error())
			}
			defer os.Remove(tempFilePath)
			if !assert.EqualFileContent(tempFilePath, origPath) {
				t.Errorf("Different file content error. Got %s is different to expected %s.", tempFilePath, origPath)
			}

			// Delete.
			err = File.Delete(file.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			err = File.Get(file.ID, "/dev/null")
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}
