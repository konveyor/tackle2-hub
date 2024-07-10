package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
)

var (
	Settings = &settings.Settings
	Log      = logr.WithName("metrics")
)

const (
	Unit = time.Second
)

// Manager provides metrics management.
type Manager struct {
	// DB
	DB *gorm.DB
}

// Run the manager.
func (m *Manager) Run(ctx context.Context, wg *sync.WaitGroup) {
	go func() {
		Log.Info("Started.")
		defer Log.Info("Died.")
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.pause()):
				time.Sleep(time.Second * 30)
				m.gaugeApplications()
			}
		}
	}()
}

func (m *Manager) pause() (d time.Duration) {
	d = Unit * time.Duration(Settings.Frequency.Metrics)
	return
}

// gaugeApplications reports the number of applications in inventory
func (m *Manager) gaugeApplications() {
	count := int64(0)
	result := m.DB.Model(&model.Application{}).Count(&count)
	if result.Error != nil {
		Log.Error(result.Error, "unable to gauge applications")
	}
	Applications.Set(float64(count))
}
