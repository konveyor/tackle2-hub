package v14

import (
	"strings"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v14/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v14")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		return
	}
	// add mvn:// prefix.
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

func (r Migration) Models() []any {
	return model.All()
}
