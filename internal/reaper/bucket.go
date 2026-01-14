package reaper

import (
	"os"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"gorm.io/gorm"
)

// BucketReaper bucket reaper.
type BucketReaper struct {
	// DB
	DB *gorm.DB
}

// Run Executes the reaper.
// A bucket is deleted when it is no longer referenced and the TTL has expired.
func (r *BucketReaper) Run() {
	Log.V(1).Info("Reaping buckets.")
	list := []model.Bucket{}
	err := r.DB.Find(&list).Error
	if err != nil {
		Log.Error(err, "")
		return
	}
	if len(list) == 0 {
		return
	}
	ids := make(map[uint]byte)
	finder := RefFinder{DB: r.DB}
	for _, m := range []any{
		&model.Application{},
		&model.TaskGroup{},
		&model.Task{},
	} {
		err := finder.Find(m, "bucket", ids)
		if err != nil {
			Log.Error(err, "")
			continue
		}
	}
	for _, bucket := range list {
		if _, found := ids[bucket.ID]; found {
			if bucket.Expiration != nil {
				bucket.Expiration = nil
				err = r.DB.Save(&bucket).Error
				Log.Error(err, "")
			}
			continue
		}
		if bucket.Expiration == nil {
			Log.Info("Bucket (orphan) found.", "id", bucket.ID, "path", bucket.Path)
			mark := time.Now().Add(time.Minute * time.Duration(Settings.Bucket.TTL))
			bucket.Expiration = &mark
			err = r.DB.Save(&bucket).Error
			Log.Error(err, "")
			continue
		}
		mark := time.Now()
		if mark.After(*bucket.Expiration) {
			err = r.delete(&bucket)
			if err != nil {
				Log.Error(err, "")
				continue
			}
		}
	}
}

// Delete bucket.
func (r *BucketReaper) delete(bucket *model.Bucket) (err error) {
	err = nas.RmDir(bucket.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			err = liberr.Wrap(
				err,
				"id",
				bucket.ID,
				"path",
				bucket.Path)
			return
		} else {
			err = nil
		}
	}
	err = r.DB.Delete(bucket).Error
	if err != nil {
		err = liberr.Wrap(
			err,
			"id",
			bucket.ID,
			"path",
			bucket.Path)
		return
	}
	Log.Info("Bucket (orphan) deleted.", "id", bucket.ID, "path", bucket.Path)
	return
}
