package reaper

import (
	"context"
	"time"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/heap"
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

func init() {
	Log = Log.V(Settings.Log.Reaper)
}

type Task = task.Task

// Manager provides task management.
type Manager struct {
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	if Settings.Log.Reaper > 0 {
		m.DB = m.DB.Debug()
	}
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
		for {
			select {
			case <-ctx.Done():
				return
			default:
				mark := time.Now()
				for _, r := range registered {
					r.Run()
				}
				Log.Info("Duration: " + time.Since(mark).String())
				heap.Free()
				m.pause()
			}
		}
	}()
}

// Pause.
func (m *Manager) pause() {
	d := Unit * time.Duration(Settings.Frequency.Reaper)
	time.Sleep(d)
}

// Reaper interface.
type Reaper interface {
	Run()
}
