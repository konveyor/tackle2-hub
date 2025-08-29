package seed

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JSON = model.JSON

type Setting struct {
}

// Apply seeds the settings.
func (r *Setting) Apply(db *gorm.DB) (err error) {
	db = db.Clauses(clause.OnConflict{
		DoNothing: true,
	})
	settings := []model.Setting{
		{Key: "git.insecure.enabled", Value: JSON("false")},
		{Key: "svn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.dependencies.update.forced", Value: JSON("false")},
	}
	for _, m := range settings {
		err = db.Create(&m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	proxies := []model.Proxy{
		{Kind: "http", Host: "", Port: 0},
		{Kind: "https", Host: "", Port: 0},
	}
	for _, m := range proxies {
		err = db.Save(&m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}
