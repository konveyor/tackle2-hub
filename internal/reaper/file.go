package reaper

import (
	"os"
	"time"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
)

// FileReaper file reaper.
type FileReaper struct {
	// DB
	DB *gorm.DB
}

// Run Executes the reaper.
// A file is deleted when it is no longer referenced and the TTL has expired.
func (r *FileReaper) Run() {
	Log.V(1).Info("Reaping files.")
	list := []model.File{}
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
		&model.Task{},
		&model.TaskReport{},
		&model.Rule{},
		&model.Target{},
		&model.AnalysisProfile{},
	} {
		err := finder.Find(m, "file", ids)
		if err != nil {
			Log.Error(err, "")
			continue
		}
	}
	for _, file := range list {
		if _, found := ids[file.ID]; found {
			if file.Expiration != nil {
				file.Expiration = nil
				err = r.DB.Save(&file).Error
				Log.Error(err, "")
			}
			continue
		}
		if file.Expiration == nil {
			Log.Info("File (orphan) found.", "id", file.ID, "path", file.Path)
			mark := time.Now().Add(time.Minute * time.Duration(Settings.File.TTL))
			file.Expiration = &mark
			err = r.DB.Save(&file).Error
			Log.Error(err, "")
			continue
		}
		mark := time.Now()
		if mark.After(*file.Expiration) {
			err = r.delete(&file)
			if err != nil {
				Log.Error(err, "")
				continue
			}
		}
	}
}

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
	Log.Info("File (orphan) deleted.", "id", file.ID, "path", file.Path)
	return
}
