package bucket

import (
	"testing"

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
