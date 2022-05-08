package volume

import (
	"errors"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	Unit = time.Minute
)

var (
	Settings = &settings.Settings
	Log      = logging.WithName("volume-scheduler")
)

//
// Manager provides task management.
type Manager struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
	// Current taskID
	taskID uint
}

//
// Run the manager.
func (m *Manager) Run(trigger chan int) {
	d := Unit * time.Duration(
		Settings.Frequency.Volume)
	go func() {
		Log.Info("Started.")
		defer Log.Info("Done.")
		for {
			select {
			case _, open := <-trigger:
				if open {
					m.update()
				} else {
					return
				}
			case <-time.After(d):
				m.update()
			}
		}
	}()
}

//
// update volumes.
// The last task is deleted and the task ID reused.
func (m *Manager) update() {
	list := []model.Volume{}
	err := m.DB.Find(&list).Error
	if err != nil {
		Log.Trace(err)
		return
	}
	ids := []uint{}
	for _, v := range list {
		ids = append(
			ids, v.ID)
	}
	r := api.Task{}
	r.Variant = "mount:report"
	r.Name = r.Variant
	r.Locator = r.Variant
	r.Addon = "admin"
	r.State = task.Ready
	r.Priority = 1
	r.TTL = m.ttl()
	r.Data = map[string]interface{}{
		"volumes": ids,
	}
	t := r.Model()
	err = m.endTask()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			Log.Trace(err)
			return
		}
	}
	t.ID = m.taskID
	err = m.DB.Create(t).Error
	if err != nil {
		Log.Trace(err)
		m.taskID = 0
	} else {
		m.taskID = t.ID
	}
}

//
// endTask ends the current task.
func (m *Manager) endTask() (err error) {
	defer func() {
		if err != nil {
			m.taskID = 0
		}
	}()
	if m.taskID == 0 {
		return
	}
	current := &model.Task{}
	current.ID = m.taskID
	err = m.DB.First(current).Error
	if err != nil {
		return
	}
	rt := task.Task{Task: current}
	err = rt.Delete(m.Client)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	db := m.DB.Select(clause.Associations)
	err = db.Delete(current).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	db = m.DB.Select(clause.Associations)
	err = db.Delete(current).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

//
// Task ttl.
func (m *Manager) ttl() (ttl *api.TTL) {
	freq := Settings.Frequency.Volume
	return &api.TTL{
		Pending:   2 * freq,
		Succeeded: 2 * freq,
		Failed:    4 * freq,
	}
}
