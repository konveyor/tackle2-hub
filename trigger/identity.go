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
	for i := range m.Applications {
		err = tr.Updated(&m.Applications[i])
		if err != nil {
			return
		}
	}
	return
}
