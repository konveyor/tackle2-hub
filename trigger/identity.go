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
	err = r.Updated(m)
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
	var appList []model.Application
	err = r.DB.Find(&appList, affected).Error
	if err != nil {
		return
	}
	for i := range appList {
		err = tr.Updated(&m.Applications[i])
		if err != nil {
			return
		}
	}
	return
}

// affected returns a list of affected application ids.
// An application is affected when:
//- the changed identity is directly associated
//- the changed identity is the default (for the kind) and an
//  application does not have an identity of the same kind
//  directly associated.
func (r *Identity) affected(changed *model.Identity) (appIds []uint, err error) {
	type M struct {
		ID      uint
		Kind    string
		Default bool
		AppId   uint
	}
	db := r.DB.Select(
		"i.ID ID",
		"i.Kind Kind",
		"i.`Default` `Default`",
		"j.ApplicationID AppId")
	db = db.Table("Identity i")
	db = db.Joins("LEFT JOIN ApplicationIdentity j ON i.ID = j.IdentityID")
	db = db.Where("i.Kind", changed.Kind)
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
		if m2.AppId > 0 {
			direct[m2.AppId] = m2.ID
		}
	}
	indirect := make(map[uint]uint)
	for _, m2 := range records {
		if m2.Default {
			_, hasDirect := direct[m2.AppId]
			if hasDirect || m2.AppId == 0 {
				continue
			}
			indirect[m2.AppId] = m2.ID
		}
	}
	for appId, _ := range direct {
		appIds = append(appIds, appId)
	}
	for appId, _ := range indirect {
		appIds = append(appIds, appId)
	}
	return
}
