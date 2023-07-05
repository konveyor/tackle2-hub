package bucket

import (
	"testing"
	"time"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestBucketCRUD(t *testing.T) {
	for _, sample := range Samples {
		t.Run(sample.Path, func(t *testing.T) {

			expectedBucket := api.Bucket{
				Path: sample.Path,
			}

			// Create a new bucket.
			assert.Must(t, Bucket.Create(&expectedBucket))

			// Get all the buckets.
			_, err := Bucket.List()
			if err != nil {
				t.Errorf(err.Error())
			}

			// Inject Expected Buckets's ID into the BucketRoot.
			pathForBucket := binding.Path(api.BucketRoot).Inject(binding.Params{api.ID: expectedBucket.ID})

			// Get specific buckets
			gotBucket := api.Bucket{}
			err = Client.Get(pathForBucket, &gotBucket)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare bucket Paths
			if gotBucket.Path != expectedBucket.Path {
				t.Errorf("Difference in Path between the buckets %v and %v", gotBucket.Path, expectedBucket.Path)
			}

			// Delete bucket.
			assert.Must(t, Bucket.Delete(pathForBucket))
		})
	}
}

// func TestBucketFile(t *testing.T) {
// 	for _, sample := range Samples {
// 		t.Run("Bucket File Test", func(t *testing.T) {
// 			err := Client.BucketPut(sample.Path, api.BucketContentRoot)
// 			if err != nil {
// 				t.Errorf(err.Error())
// 			}
// 		})
// 	}
// }

func TestBucketReaper(t *testing.T) {
	for _, sample := range ReaperSamples {
		t.Run("BucketReaper", func(t *testing.T) {
			expectedReaperBucket := api.Bucket{
				Path:       sample.Path,
				Expiration: sample.Expiration,
			}

			assert.Must(t, Bucket.Create(&expectedReaperBucket))

			// Inject Reaper Buckets's ID into the BucketRoot.
			pathForReaperBucket := binding.Path(api.BucketRoot).Inject(binding.Params{api.ID: expectedReaperBucket.ID})

			gotReaperBucket := api.Bucket{}
			for {
				time.Sleep(time.Second * 10)
				err := Client.Get(pathForReaperBucket, &gotReaperBucket)
				if err != nil {
					break
				}
			}
		})
	}
}
