package trigger

import (
	"github.com/konveyor/tackle2-hub/model"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
)

// Identity trigger.
type Identity struct {
	Trigger
	TaskManager *tasking.Manager
	DB          *gorm.DB
}

// Updated model created trigger.
func (r *Identity) Updated(m *model.Identity) (err error) {
	tr := Application{
		TaskManager: r.TaskManager,
		DB:          r.DB,
	}
	for i := range m.Applications {
		err = tr.Updated(&m.Applications[i])
		if err != nil {
			return
		}
	}
	return
}
