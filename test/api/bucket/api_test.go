package bucket

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestBucketCRUD(t *testing.T) {
	for _, sample := range Samples {
		t.Run("Bucket CRUD Test", func(t *testing.T) {

			expectedBucket := api.Bucket{
				Path: sample.Path,
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
		})
	}
}

func TestBucketFile(t *testing.T) {
	for _, sample := range Samples {
		t.Run("Bucket File Test", func(t *testing.T) {
			// Create a sample bucket.
			expectedBucket := api.Bucket{
				Path: sample.Path,
			}
			// Create a test bucket.
			assert.Must(t, Bucket.Create(&expectedBucket))

			// Inject bucket id and location into the path
			bucketContentPath := binding.Path(api.BucketContentRoot).Inject(binding.Params{api.ID: expectedBucket.ID, api.Wildcard: sample.Path})

			// Add the file to the Bucket.
			assert.Must(t, Client.BucketPut(sample.Path, bucketContentPath))

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
			expectedCSV, err := ioutil.ReadFile(sample.Path)
			if err != nil {
				t.Errorf("Error reading CSV: %s", sample.Path)
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
		})
	}
}

func TestBucketReaper(t *testing.T) {
	for _, sample := range ReaperSamples {
		t.Run("Indirect Bucket Deletion Check", func(t *testing.T) {
			expectedReaperBucket := api.Bucket{
				Path:       sample.Path,
				Expiration: sample.Expiration,
			}

			// Create a new Bucket.
			assert.Must(t, Bucket.Create(&expectedReaperBucket))

			// Inject Reaper Buckets's ID into the BucketRoot.
			pathForReaperBucket := binding.Path(api.BucketRoot).Inject(binding.Params{api.ID: expectedReaperBucket.ID})

			// Wait for Bucket to be deleted.
			gotReaperBucket := api.Bucket{}
			for {
				time.Sleep(time.Second * 10)
				err := Client.Get(pathForReaperBucket, &gotReaperBucket)
				// error will not be nil when the bucket is deleted thus will break the loop.
				if err != nil {
					break
				}
			}
		})
	}
}
