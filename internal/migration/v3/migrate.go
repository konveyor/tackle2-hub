package v3

import (
	"encoding/json"

	liberr "github.com/jortel/go-utils/error"
	model3 "github.com/konveyor/tackle2-hub/internal/migration/v2/model"
	model2 "github.com/konveyor/tackle2-hub/internal/migration/v3/model"
	"github.com/konveyor/tackle2-hub/internal/migration/v3/seed"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	//
	// Tags/Categories.
	err = db.Migrator().RenameTable(model3.TagType{}, model2.TagCategory{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Migrator().RenameColumn(model2.Tag{}, "TagTypeID", "CategoryID")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Migrator().RenameColumn(model2.ImportTag{}, "TagType", "Category")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	//
	// Facts.
	err = r.factMigration(db)
	if err != nil {
		return
	}
	//
	// Buckets.
	err = r.bucketMigration(db)
	if err != nil {
		return
	}
	//
	// Altering the primary key requires constructing a new table, so rename the old one,
	// create the new one, copy over the rows, and then drop the old one.
	err = db.Migrator().RenameTable("ApplicationTags", "ApplicationTags__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Migrator().CreateTable(model2.ApplicationTag{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	result := db.Exec("INSERT INTO ApplicationTags (ApplicationID, TagID, Source) SELECT ApplicationID, TagID, '' FROM ApplicationTags__old;")
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	err = db.Migrator().DropTable("ApplicationTags__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	//
	// Models.
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	//
	// Seed.
	seed.Seed(db)

	return
}

func (r Migration) Models() []any {
	return model2.All()
}

// factMigration migrates Application.Facts.
// This involves changing the Facts type from JSON which maps to
// a column in the DB to an ORM virtual field. This, and the data
// migration both require the v2 model.
func (r Migration) factMigration(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	list := []model3.Application{}
	result := db.Find(&list)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	err = migrator.AutoMigrate(&model2.Fact{})
	if err != nil {
		return
	}
	for _, m := range list {
		d := map[string]any{}
		_ = json.Unmarshal(m.Facts, &d)
		for k, v := range d {
			jv, _ := json.Marshal(v)
			fact := &model2.Fact{}
			fact.ApplicationID = m.ID
			fact.Key = k
			fact.Value = jv
			result := db.Create(fact)
			if result.Error != nil {
				err = liberr.Wrap(result.Error)
				return
			}
		}
	}
	err = migrator.DropColumn(&model3.Application{}, "Facts")
	if err != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	return
}

// bucketMigration migrates buckets.
func (r Migration) bucketMigration(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	err = migrator.AutoMigrate(&model2.Bucket{})
	if err != nil {
		return
	}
	err = r.appBucketMigration(db)
	if err != nil {
		return
	}
	err = r.taskBucketMigration(db)
	if err != nil {
		return
	}
	err = r.taskGroupBucketMigration(db)
	if err != nil {
		return
	}

	return
}

// appBucketMigration migrates application buckets.
// The (v2) Application.Bucket (string) contains the bucket storage path. Migration needs to
// build a `Bucket` object using this path for each and set v3 BucketID.
// The Application.Bucket becomes virtual.
func (r Migration) appBucketMigration(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	err = migrator.AutoMigrate(&model2.Application{})
	if err != nil {
		return
	}
	list := []model3.Application{}
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		if m.Bucket == "" {
			continue
		}
		bucket := &model2.Bucket{}
		bucket.Path = m.Bucket
		err = db.Create(bucket).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		db := db.Model(&model2.Application{})
		db = db.Where("ID = ?", m.ID)
		result := db.Update("BucketID", &bucket.ID)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	err = migrator.DropColumn(&model3.Application{}, "Bucket")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

// taskBucketMigration migrates task buckets.
// The (v2) Task.Bucket (string) contains the bucket storage path. Migration needs to
// build a `Bucket` object using this path for each and set v3 BucketID.
// The Task.Bucket becomes virtual.
func (r Migration) taskBucketMigration(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	err = migrator.AutoMigrate(&model2.Task{})
	if err != nil {
		return
	}
	list := []model3.Task{}
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		if m.Bucket == "" {
			continue
		}
		bucket := &model2.Bucket{}
		bucket.Path = m.Bucket
		err = db.Create(bucket).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		db := db.Model(&model2.Task{})
		db = db.Where("ID = ?", m.ID)
		result := db.Update("BucketID", &bucket.ID)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	err = migrator.DropColumn(&model3.Task{}, "Bucket")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

// taskGroupBucketMigration migrates task group buckets.
// The (v2) TaskGroup.Bucket (string) contains the bucket storage path. Migration needs to
// build a `Bucket` object using this path for each and set v3 BucketID.
// The TaskGroup.Bucket becomes virtual.
func (r Migration) taskGroupBucketMigration(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	err = migrator.AutoMigrate(&model2.TaskGroup{})
	if err != nil {
		return
	}
	list := []model3.TaskGroup{}
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		if m.Bucket == "" {
			continue
		}
		bucket := &model2.Bucket{}
		bucket.Path = m.Bucket
		err = db.Create(bucket).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		db := db.Model(&model2.TaskGroup{})
		db = db.Where("ID = ?", m.ID)
		result := db.Update("BucketID", &bucket.ID)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	err = migrator.DropColumn(&model3.TaskGroup{}, "Bucket")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}
