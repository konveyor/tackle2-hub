package v23

import (
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/migration/json"
	v22 "github.com/konveyor/tackle2-hub/internal/migration/v22/model"
	"github.com/konveyor/tackle2-hub/internal/migration/v23/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = r.encodeScopes(db)
	if err != nil {
		return
	}
	err = db.AutoMigrate(r.Models()...)
	return
}

// encodeScopes json encodes scopes stored
// as space-delimited as json array.
// The Token.Scopes in type changed to []string.
func (r Migration) encodeScopes(db *gorm.DB) (err error) {
	var list []*v22.Token
	err = db.Find(&list).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, m := range list {
		if m.Scopes == "" {
			continue
		}
		if m.Scopes[0] == '[' {
			continue
		}
		var b []byte
		v := strings.Fields(m.Scopes)
		b, err = json.Marshal(v)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		m.Scopes = string(b)
		err = db.Save(m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}

func (r Migration) Models() []any {
	return model.All()
}
