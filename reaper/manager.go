package reaper

import (
	"context"
	"sync"
	"time"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/settings"
	"github.com/konveyor/tackle2-hub/task"
	"gorm.io/gorm"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Unit = time.Minute
)

var (
	Settings = &settings.Settings
	Log      = logr.WithName("reaper")
)

type Task = task.Task

// Manager provides task management.
type Manager struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
}

// Run the manager.
func (m *Manager) Run(ctx context.Context, wg *sync.WaitGroup) {
	registered := []Reaper{
		&TaskReaper{
			Client: m.Client,
			DB:     m.DB,
		},
		&GroupReaper{
			DB: m.DB,
		},
		&BucketReaper{
			DB: m.DB,
		},
		&FileReaper{
			DB: m.DB,
		},
	}
	go func() {
		Log.Info("Started.")
		defer Log.Info("Died.")
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(m.pause()):
				for _, r := range registered {
					r.Run()
				}
			}
		}
	}()
}

// Pause.
func (m *Manager) pause() (d time.Duration) {
	d = Unit * time.Duration(Settings.Frequency.Reaper)
	return
}

// Reaper interface.
type Reaper interface {
	Run()
}
