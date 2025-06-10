package v18

import (
	v17 "github.com/konveyor/tackle2-hub/migration/v17/model"
	"github.com/konveyor/tackle2-hub/migration/v18/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = r.renameIssueToInsight(db)
	if err != nil {
		return
	}
	err = db.AutoMigrate(r.Models()...)
	return
}

func (r Migration) Models() []any {
	return model.All()
}

func (r Migration) renameIssueToInsight(db *gorm.DB) (err error) {
	migrator := db.Migrator()
	if !migrator.HasTable(&v17.Issue{}) {
		return
	}
	err = migrator.RenameColumn(&v17.Issue{}, "issueID", "insightID")
	if err != nil {
		return
	}
	err = migrator.DropIndex(v17.Issue{}, "issueA")
	if err != nil {
		return
	}
	err = migrator.RenameTable(v17.Issue{}, model.Insight{})
	return
}
