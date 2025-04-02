package v2

import (
	"github.com/konveyor/tackle2-hub/migration/v3/model"

	"gorm.io/gorm"
)

type JSON = model.JSON

// Seed the database with models.
func seed(db *gorm.DB) {
	settings := []model.Setting{
		{Key: "git.insecure.enabled", Value: JSON("false")},
		{Key: "svn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.insecure.enabled", Value: JSON("false")},
		{Key: "mvn.dependencies.update.forced", Value: JSON("false")},
	}
	_ = db.Create(settings)
	proxies := []model.Proxy{
		{Kind: "http", Host: "", Port: 0},
		{Kind: "https", Host: "", Port: 0},
	}
	_ = db.Create(proxies)
	return
}
