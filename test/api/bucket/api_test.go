package bucket

import (
	"bytes"
	"compress/gzip"
	"io"
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
		t.Run("Bucket CRUD Test", func(t *testing.T) {

			expectedBucket := api.Bucket{
				Path: bucket.Path,
			}

			// Create a new bucket.
			assert.Must(t, Bucket.Create(&expectedBucket))

			// Get all the buckets.
			gotBuckets, err := Bucket.List()
			if err != nil {
				t.Errorf(err.Error())
			}

			// Find for the specific bucket and compare Paths as it is a unique value.
			for _, gotBucket := range gotBuckets {
				if gotBucket.ID == expectedBucket.ID {
					if gotBucket.Path != expectedBucket.Path {
						t.Errorf("Difference in Path between the buckets %v and %v", gotBucket.Path, expectedBucket.Path)
					}
				}
			}

			// Inject Expected Buckets's ID into the BucketRoot.
			pathForBucket := binding.Path(api.BucketRoot).Inject(binding.Params{api.ID: expectedBucket.ID})

			// Get specific bucket.
			gotBucket := api.Bucket{}
			err = Client.Get(pathForBucket, &gotBucket)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare bucket Paths.
			if gotBucket.Path != expectedBucket.Path {
				t.Errorf("Difference in Path between the buckets %v and %v", gotBucket.Path, expectedBucket.Path)
			}

			// Delete bucket.
			assert.Must(t, Bucket.Delete(pathForBucket))

			// Create a test bucket.
			assert.Must(t, Bucket.Create(&expectedBucket))

			// Inject bucket id and location into the path
			bucketContentPath := binding.Path(api.BucketContentRoot).Inject(binding.Params{api.ID: expectedBucket.ID, api.Wildcard: bucket.Path})

			// Add the file to the Bucket.
			assert.Must(t, Client.BucketPut(bucket.Path, bucketContentPath))

			// Get the file from the bucket.
			pathToGotCSV := "downloadcsv.csv"
			assert.Should(t, Client.BucketGet(bucketContentPath, pathToGotCSV))

			// Read the got CSV file.
			gotCSV, err := ioutil.ReadFile(pathToGotCSV)
			if err != nil {
				t.Errorf("Error reading CSV: %s", pathToGotCSV)
			}
			gotCSVString := string(gotCSV)

			// Read the expected CSV file.
			expectedCSV, err := ioutil.ReadFile(bucket.Path)
			if err != nil {
				t.Errorf("Error reading CSV: %s", bucket.Path)
			}
			expectedCSVString := string(expectedCSV)
			if gotCSVString != expectedCSVString {
				t.Errorf("The CSV files have different content %s and %s", gotCSVString, expectedCSVString)
			}

			// Remove the CSV file created.
			err = os.Remove(pathToGotCSV)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Delete the bucket contents.
			assert.Must(t, Client.Delete(bucketContentPath))

			baseDirectory := "sample"

			// Generate a unique temporary directory path
			tempDir, err := ioutil.TempDir(baseDirectory, "")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer os.RemoveAll(tempDir)

			// Construct the destination path for the CSV file in the temporary directory
			destFilePath := filepath.Join(tempDir, "template_application_import.csv")

			// Open the source CSV file for reading
			srcFile, err := os.Open(bucket.Path)
			if err != nil {
				t.Errorf(err.Error())
			}
			defer srcFile.Close()

			// Create the destination file in the temporary directory
			destFile, err := os.Create(destFilePath)
			if err != nil {
				t.Errorf(err.Error())
			}
			defer destFile.Close()

			// Copy the contents of the source CSV file to the destination file
			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				t.Errorf(err.Error())
			}

			var buf bytes.Buffer
			err = Bucket.PutDir(&buf, tempDir)
			if err != nil {
				t.Errorf(err.Error())
			}

			outputFile, err := os.Create("test.tar.gz")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer outputFile.Close()

			// Create a gzip writer
			gzipWriter := gzip.NewWriter(outputFile)
			defer gzipWriter.Close()

			// Write the archive data from the buffer to the gzip writer
			_, err = buf.WriteTo(gzipWriter)
			if err != nil {
				t.Errorf(err.Error())
			}

			// expectedFile, err := os.Open("test.tar.gz")
			// if err != nil {
			// 	t.Errorf(err.Error())
			// }
			// defer expectedFile.Close()

			// err = Bucket.GetDir(expectedFile, baseDirectory)
			// if err != nil {
			// 	t.Errorf(err.Error())
			// }

			err = os.Remove("test.tar.gz")
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
