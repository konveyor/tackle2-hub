package v18

import (
	v17 "github.com/konveyor/tackle2-hub/internal/migration/v17/model"
	"github.com/konveyor/tackle2-hub/internal/migration/v18/model"
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
	// issue
	if !migrator.HasTable(&v17.Issue{}) {
		return
	}
	err = migrator.DropIndex(&v17.Issue{}, "issueA")
	if err != nil {
		return
	}
	err = migrator.RenameTable(&v17.Issue{}, model.Insight{})
	if err != nil {
		return
	}
	// incident
	err = db.Exec("CREATE TABLE Tmp AS SELECT * FROM Incident").Error
	if err != nil {
		return
	}
	err = migrator.DropTable(&v17.Incident{})
	if err != nil {
		return
	}
	err = migrator.AutoMigrate(&model.Incident{})
	if err != nil {
		return
	}
	err = db.Exec("INSERT INTO Incident SELECT * FROM Tmp").Error
	if err != nil {
		return
	}
	err = migrator.DropTable("Tmp")
	if err != nil {
		return
	}
	return
}
