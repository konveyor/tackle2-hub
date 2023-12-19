package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
	"k8s.io/apimachinery/pkg/util/rand"
	"io/ioutil"
	"strings"
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

func TestFileTouchPatchGetDelete(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Touch.
			name := "Patch-Test"
			file, err := File.Touch(name)
			if err != nil {
				t.Errorf(err.Error())
			}
			// Patch (append)
			content := "This is my Test. "
			for _, p := range strings.Fields(content) {
				err = File.Patch(file.ID, []byte(p+" "))
				if err != nil {
					t.Errorf(err.Error())
				}
			}
			// Get.
			tmp := fmt.Sprintf(
				"/tmp/%s-%d",
				r.Name,
				rand.Int())
			err = File.Get(file.ID, tmp)
			if err != nil {
				t.Errorf(err.Error())
			}
			defer func() {
				_ = os.Remove(tmp)
			}()
			if file.Name != name {
				t.Errorf(
					"File name mismatch. Expected: '%s' found: '%s'",
					name,
					file.Name)
			}

			f, err := os.Open(tmp)
			if err != nil {
				t.Errorf(err.Error())
			}
			read, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf(err.Error())
			}
			if content != string(read) {
				t.Errorf(
					"File content mismatch. Expcected: '%s' read: '%s'",
					content,
					string(read))
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
