package bucket

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestBucketCRUD(t *testing.T) {
	for _, bucket := range Buckets {

		t.Run("Create Bucket and Compare", func(t *testing.T) {

			// Create a new bucket.
			assert.Must(t, Bucket.Create(&bucket))

			// Get all the buckets.
			gotBuckets, err := Bucket.List()
			if err != nil {
				t.Errorf(err.Error())
			}

			// Find for the specific bucket and compare Paths as it is a unique value.
			for _, gotBucket := range gotBuckets {
				if gotBucket.ID == bucket.ID {
					if gotBucket.Path != bucket.Path {
						t.Errorf("Difference in Path between the buckets %v and %v", gotBucket.Path, bucket.Path)
					}
				}
			}
			fmt.Println("hello")

			// Inject Expected Buckets's ID into the BucketRoot.
			bucketID := binding.Path(api.BucketRoot).Inject(binding.Params{api.ID: bucket.ID})

			// Get specific bucket.
			gotBucket := api.Bucket{}
			err = Client.Get(bucketID, &gotBucket)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare bucket Paths.
			if gotBucket.Path != bucket.Path {
				t.Errorf("Difference in Path between the buckets %v and %v", gotBucket.Path, bucket.Path)
			}
		})

		t.Run("File and Directory Tests", func(t *testing.T) {
			expectedBucket := Bucket.Content(bucket.ID)

			expectedFile, err := ioutil.TempFile("", "a")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer os.Remove(expectedFile.Name())

			data := []byte("Hello World")
			_, err = expectedFile.Write(data)
			if err != nil {
				t.Errorf(err.Error())
			}

			assert.Should(t, expectedBucket.Put(expectedFile.Name(), expectedFile.Name()))

			gotFile, err := ioutil.TempFile("", "b")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer os.Remove(gotFile.Name())

			assert.Should(t, expectedBucket.Get(expectedFile.Name(), gotFile.Name()))

			expected, err := ioutil.ReadFile(expectedFile.Name())
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err := ioutil.ReadFile(gotFile.Name())
			if err != nil {
				t.Errorf(err.Error())
			}

			if len(expected) != len(got) {
				t.Errorf("Mismatch in outputs")
			}

			/*----------------------Directory Tests----------------------*/

			// Create a sample directory
			expectedDir, err := ioutil.TempDir("", "tree")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer os.RemoveAll(expectedDir)

			// Create and write data to text files in the temporary directory
			for i := 1; i <= 5; i++ {
				fileName := fmt.Sprintf("file%d.txt", i)
				filePath := filepath.Join(expectedDir, fileName)

				// Create a new text file
				file, err := os.Create(filePath)
				if err != nil {
					t.Errorf(err.Error())
				}
				defer file.Close()

				// Write data to the file
				data := []byte("Hello world!")
				_, err = file.Write(data)
				if err != nil {
					t.Errorf(err.Error())
				}
			}

			assert.Should(t, expectedBucket.Put(expectedDir, expectedDir))

			gotDir, err := ioutil.TempDir("", "b")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer os.RemoveAll(gotDir)

			assert.Should(t, expectedBucket.Get(expectedDir, gotDir))

			expectedDirContent, err := ioutil.ReadDir(expectedDir)
			if err != nil {
				t.Errorf(err.Error())
			}

			gotDirContent, err := ioutil.ReadDir(gotDir)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare length of expected and got Directory content
			if len(expectedDirContent) != len(gotDirContent) {
				t.Errorf("Mismatch in outputs")
			}

			// Compare elementwise content of expected and got Directory
			for i := 0; i < len(expectedDirContent); i++ {
				expectedDirInfo := expectedDirContent[i]
				gotDirInfo := gotDirContent[i]

				if expectedDirInfo.Name() != gotDirInfo.Name() {
					t.Errorf("Mismatch in names expected: %v, got: %v", expectedDirInfo.Name(), gotDirInfo.Name())
				}

				if expectedDirInfo.Size() != gotDirInfo.Size() {
					t.Errorf("Mismatch in sizes expected: %v, got %v", expectedDirInfo.Size(), gotDirInfo.Size())
				}

				if expectedDirInfo.Mode() != gotDirInfo.Mode() {
					t.Errorf("Mismatch in modes expected: %v, got %v", expectedDirInfo.Mode(), gotDirInfo.Mode())
				}

			}
		})

		// Delete the bucket
		assert.Must(t, Bucket.Delete(bucket.ID))
	}
}
