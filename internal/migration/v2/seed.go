package v2

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/migration/v3/model"
	"gorm.io/gorm"
)

type JSON = model.JSON

// Seed the database with models.
func seed(db *gorm.DB) (err error) {
	settings := []model.Setting{
		{Key: "git.insecure.enabled", Value: JSON("false")},
		{Key: "svn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.dependencies.update.forced", Value: JSON("false")},
	}
	err = db.Create(settings).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	proxies := []model.Proxy{
		{Kind: "http", Host: "", Port: 0},
		{Kind: "https", Host: "", Port: 0},
	}
	err = db.Create(proxies).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
