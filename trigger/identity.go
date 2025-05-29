package trigger

import (
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

// Identity trigger.
type Identity struct {
	Trigger
}

// Updated model created trigger.
func (r *Identity) Updated(m *model.Identity) (err error) {
	tr := Application{
		Trigger: Trigger{
			TaskManager: r.TaskManager,
			Client:      r.Client,
			DB:          r.DB,
		},
	}
	id := m.ID
	m = &model.Identity{}
	db := r.DB.Preload(clause.Associations)
	err = db.First(m, id).Error
	if err != nil {
		return
	}
	if m.Default {
		direct := make(map[uint]byte)
		for i := range m.Applications {
			appId := m.Applications[i].ID
			direct[appId] = 0
		}
		type M struct {
			AppId uint
			ID    uint
			Kind  string
		}
		db := r.DB.Select(
			"j.ApplicationID AppId",
			"i.ID ID",
			"i.Kind Kind")
		db = r.DB.Table("Identity i")
		db = db.Joins("JOIN ApplicationIdentity j ON a.ID = j.IdentityID")
		err = r.DB.Find(&applications).Error
		if err != nil {
			return
		}
		for i := range m.Applications {
			app := &m.Applications[i]
		}
	}
	for i := range m.Applications {
		err = tr.Updated(&m.Applications[i])
		if err != nil {
			return
		}
	}
	return
}

func (r *Identity) affected(m *model.Identity) (appIds []uint, err error) {
	type M struct {
		ID      uint
		Kind    string
		Default bool
		AppId   uint
	}
	var identities []M
	db := r.DB.Select(
		"i.ID ID",
		"i.Kind Kind",
		"i.Default Default",
		"j.ApplicationID AppId")
	db = r.DB.Table("Identity i")
	db = db.Joins("JOIN ApplicationIdentity j ON a.ID = j.IdentityID")
	db = db.Where("i.Kind", m.Kind)
	err = r.DB.Find(&identities).Error
	if err != nil {
		return
	}
	direct := make(map[uint]byte)
	for i := range identities {
		m2 := &identities[i]
		if m2.AppId > 0 {
			direct[m2.ID] = 0
		}
	}
	return
}
