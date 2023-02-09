package reaper

import (
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"os"
	"time"
)

//
// FileReaper file reaper.
type FileReaper struct {
	// DB
	DB *gorm.DB
}

//
// Run Executes the reaper.
// A file is deleted when it is no longer referenced and the TTL has expired.
func (r *FileReaper) Run() {
	Log.V(1).Info("Reaping files.")
	list := []model.File{}
	err := r.DB.Find(&list).Error
	if err != nil {
		Log.Trace(err)
		return
	}
	for _, file := range list {
		busy, err := r.busy(&file)
		if err != nil {
			Log.Trace(err)
			continue
		}
		if busy {
			if file.Expiration != nil {
				file.Expiration = nil
				err = r.DB.Save(&file).Error
				Log.Trace(err)
			}
			continue
		}
		if file.Expiration == nil {
			mark := time.Now().Add(time.Minute * time.Duration(Settings.File.TTL))
			file.Expiration = &mark
			err = r.DB.Save(&file).Error
			Log.Trace(err)
			continue
		}
		mark := time.Now()
		if mark.After(*file.Expiration) {
			err = r.delete(&file)
			if err != nil {
				Log.Trace(err)
				continue
			}
		}
	}
}

//
// busy determines if anything references the file.
func (r *FileReaper) busy(file *model.File) (busy bool, err error) {
	nRef := int64(0)
	var n int64
	ref := RefCounter{DB: r.DB}
	for _, m := range []interface{}{
		// ADD MODELS HERE
	} {
		n, err = ref.Count(m, "file", file.ID)
		if err != nil {
			Log.Trace(err)
			continue
		}
		nRef += n
	}
	busy = nRef > 0
	return
}

//
// Delete file.
func (r *FileReaper) delete(file *model.File) (err error) {
	err = os.Remove(file.Path)
	if err != nil {
		if !os.IsNotExist(err) {
			err = liberr.Wrap(
				err,
				"id",
				file.ID,
				"path",
				file.Path)
			return
		} else {
			err = nil
		}
	}
	err = r.DB.Delete(file).Error
	if err != nil {
		err = liberr.Wrap(
			err,
			"id",
			file.ID,
			"path",
			file.Path)
		return
	}
	Log.Info("File deleted.", "id", file.ID, "path", file.Path)
	return
}
