package v14

import (
	"strings"

	"github.com/konveyor/tackle2-hub/internal/migration/v14/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		return
	}
	err = r.mavenPrefix(db)
	if err != nil {
		return
	}
	err = r.taskKind(db)
	if err != nil {
		return
	}
	return
}

// mavenPrefix ensures the Application.Binary which are maven
// coordinates have the mvn:// prefix added in 0.5.
func (r Migration) mavenPrefix(db *gorm.DB) (err error) {
	prefix := "mvn://"
	var list []*model.Application
	err = db.Find(&list).Error
	if err != nil {
		return
	}
	for _, m := range list {
		if m.Binary == "" || strings.HasPrefix(m.Binary, prefix) {
			continue
		}
		m.Binary = prefix + m.Binary
		err = db.Save(m).Error
		if err != nil {
			return
		}
	}
	return
}

// taskKind ensures tasks have a kind.
// In 0.5 task (kinds) added. A task named `analyzer` is
// installed by the operator.
func (r Migration) taskKind(db *gorm.DB) (err error) {
	kind := "analyzer"
	var list []*model.Task
	err = db.Find(&list).Error
	if err != nil {
		return
	}
	for _, m := range list {
		if m.Addon == kind && m.Kind == "" {
			m.Kind = kind
		}
		err = db.Save(m).Error
		if err != nil {
			return
		}
	}
	return
}

func (r Migration) Models() []any {
	return model.All()
}
