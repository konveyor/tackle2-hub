package bucket

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestBucket(t *testing.T) {
	for _, bucket := range Buckets {
		t.Run("Bucket CRUD Test", func(t *testing.T) {
			expectedPath := bucket.Path
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

			// Inject Expected Buckets's ID into the BucketRoot.
			pathForBucket := binding.Path(api.BucketRoot).Inject(binding.Params{api.ID: bucket.ID})

			// Get specific bucket.
			gotBucket := api.Bucket{}
			err = Client.Get(pathForBucket, &gotBucket)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare bucket Paths.
			if gotBucket.Path != bucket.Path {
				t.Errorf("Difference in Path between the buckets %v and %v", gotBucket.Path, bucket.Path)
			}

			/* -----------------------------------------------------------------------------*/
			// Directory tests

			// Inject bucket id and location into the path
			bucketContentPath := binding.Path(api.BucketContentRoot).Inject(binding.Params{api.ID: bucket.ID, api.Wildcard: expectedPath})

			// Add the file to the Bucket.
			assert.Should(t, Client.BucketPut(expectedPath, bucketContentPath))

			// Get the file from the bucket.
			pathToGotCSV := "downloadedcsv.csv"
			_, err = os.Create(pathToGotCSV)
			if err != nil {
				t.Errorf(err.Error())
			}

			assert.Should(t, Client.BucketGet(bucketContentPath, pathToGotCSV))

			// Read the got CSV file.
			gotCSV, err := ioutil.ReadFile(pathToGotCSV)
			if err != nil {
				t.Errorf("Error reading CSV: %s", pathToGotCSV)
			}
			gotCSVString := string(gotCSV)
			// Read the expected CSV file.
			expectedCSV, err := ioutil.ReadFile(expectedPath)
			if err != nil {
				t.Errorf("Error reading CSV: %s", expectedPath)
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

			assert.Should(t, Client.Delete(bucketContentPath))

			/* -----------------------------------------------------------------------------*/
			// Archive tests

			outputDirectory := "sample"

			// Generate a unique temporary directory path
			tempDir, err := ioutil.TempDir(outputDirectory, "")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer os.RemoveAll(tempDir)

			// Construct the destination path for the CSV file in the temporary directory
			destFilePath := filepath.Join(tempDir, strings.TrimPrefix(expectedPath, "sample/"))

			// Open the source CSV file for reading
			srcFile, err := os.Open(expectedPath)
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

			// Access the directory , convert to archive and upload its contents.
			var buf bytes.Buffer
			assert.Should(t, Bucket.PutDir(&buf, tempDir))

			// Create an archive.
			var outputBuffer bytes.Buffer
			_ = Bucket.Compress(expectedPath, &outputBuffer)

			// write the .tar.gzip
			fileToWrite, err := os.OpenFile("./compress.tar.gzip", os.O_CREATE|os.O_RDWR, os.FileMode(0777))
			if err != nil {
				t.Errorf(err.Error())
			}
			defer fileToWrite.Close()

			// Copy the csv file to the archive.
			_, err = io.Copy(fileToWrite, &outputBuffer)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Open the file for reading.
			expectedArchive, err := os.Open("compress.tar.gzip")
			if err != nil {
				t.Errorf(err.Error())
			}
			defer expectedArchive.Close()

			// Create the "compress" directory
			err = os.MkdirAll("compress", 0755)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Create the "sample" subdirectory inside "compress"
			err = os.MkdirAll(filepath.Join("compress", "sample"), 0755)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Extract the contents in compress/sample directory.
			assert.Should(t, Bucket.GetDir(expectedArchive, "compress"))

			gotPath := filepath.Join("compress", expectedPath)

			// Open got file for comparison.
			gotOutputCSV, err := ioutil.ReadFile(gotPath)
			if err != nil {
				t.Errorf(err.Error())
			}
			gotOutputCSVString := string(gotOutputCSV)

			// Open expected file for comparison.
			expectedOutputCSV, err := ioutil.ReadFile(expectedPath)
			if err != nil {
				t.Errorf(err.Error())
			}
			expectedOutputCSVString := string(expectedOutputCSV)

			// Compare the two csv files.
			if gotOutputCSVString != expectedOutputCSVString {
				t.Errorf("The CSV files have different content %s and %s", gotOutputCSVString, expectedOutputCSVString)
			}

			// Remove the archive.
			err = os.Remove("compress.tar.gzip")
			if err != nil {
				t.Errorf(err.Error())
			}

			// Remove the compress directory.
			err = os.RemoveAll("compress")
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
