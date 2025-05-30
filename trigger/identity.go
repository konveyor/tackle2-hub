package trigger

import (
	"github.com/konveyor/tackle2-hub/model"
)

// Identity trigger.
type Identity struct {
	Trigger
}

// Created model created trigger.
func (r *Identity) Created(m *model.Identity) (err error) {
	if m.Default {
		err = r.Updated(m)
	}
	return
}

// Updated model updated trigger.
func (r *Identity) Updated(m *model.Identity) (err error) {
	tr := Application{
		Trigger: Trigger{
			TaskManager: r.TaskManager,
			Client:      r.Client,
			DB:          r.DB,
		},
	}
	affected, err := r.affected(m)
	if err != nil {
		return
	}
	for _, batch := range affected {
		var appList []model.Application
		err = r.DB.Find(&appList, batch).Error
		if err != nil {
			return
		}
		for i := range appList {
			err = tr.Updated(&appList[i])
			if err != nil {
				return
			}
		}
	}
	return
}

// affected returns the affected application ids.
// An application is affected when:
//- the changed identity is directly associated
//- the changed identity is the default (for the kind) and an
//  application does not have an identity of the same kind
//  directly associated.
func (r *Identity) affected(changed *model.Identity) (appIds [][]uint, err error) {
	type M struct {
		AppId   uint
		Id      uint
		Kind    string
		Default bool
	}
	db := r.DB.Select(
		"a.ID AppId",
		"i.ID Id",
		"i.Kind Kind",
		"i.`Default` `Default`")
	db = db.Table("Application a")
	db = db.Joins("LEFT JOIN ApplicationIdentity j ON j.ApplicationID = a.ID")
	db = db.Joins("LEFT JOIN Identity i ON i.ID = j.IdentityID AND i.Kind = ?", changed.Kind)
	cursor, err := db.Rows()
	if err != nil {
		return
	}
	defer func() {
		_ = cursor.Close()
	}()
	m := M{}
	var records []M
	for cursor.Next() {
		err = db.ScanRows(cursor, &m)
		if err != nil {
			return
		}
		records = append(records, m)
	}
	direct := make(map[uint]uint)
	for _, m2 := range records {
		if m2.Id > 0 {
			direct[m2.AppId] = m2.Id
		}
	}
	indirect := make(map[uint]uint)
	if changed.Default {
		for _, m2 := range records {
			_, hasDirect := direct[m2.AppId]
			if !hasDirect {
				indirect[m2.AppId] = m2.Id
			}
		}
	}
	batch := make([]uint, 0, 100)
	add := func(id uint) {
		if len(batch) == cap(batch) {
			appIds = append(appIds, batch)
			batch = make([]uint, 0, cap(batch))
		}
		batch = append(batch, id)
	}
	defer func() {
		if len(batch) > 0 {
			appIds = append(appIds, batch)
		}
	}()
	for appId, _ := range direct {
		add(appId)
	}
	for appId, _ := range indirect {
		add(appId)
	}
	return
}
