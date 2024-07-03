package trigger

import (
	"github.com/konveyor/tackle2-hub/model"
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
	for i := range m.Applications {
		err = tr.Updated(&m.Applications[i])
		if err != nil {
			return
		}
	}
	return
}
