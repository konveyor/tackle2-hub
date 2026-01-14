package metrics

import (
	"context"
	"time"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
)

var (
	Log = logr.New("metrics", 0)
)

// Manager provides metrics management.
type Manager struct {
	// DB
	DB *gorm.DB
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	go func() {
		Log.Info("Started.")
		defer Log.Info("Died.")
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second * 30)
				m.gaugeApplications()
			}
		}
	}()
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
