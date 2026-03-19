package reaper

import (
	"context"
	"sync"
	"time"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/heap"
	"github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gorm.io/gorm"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Settings = &settings.Settings
	Log      = logr.New("reaper", Settings.Log.Reaper)
)

type Task = task.Task

// Manager provides task management.
type Manager struct {
	mutex sync.RWMutex
	// DB
	DB *gorm.DB
	// k8s client.
	Client k8s.Client
	// background manager
	background bool
}

// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.background {
		return
	}
	if Log.V(1).Enabled() {
		m.DB = m.DB.Debug()
	}
	threshold := 10 * time.Second
	go func() {
		defer func() {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			Log.Info("Manager stopped.")
			m.background = false
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				mark := time.Now()
				m.Iterate()
				d := time.Since(mark)
				if d > threshold {
					Log.Info("Duration: " + d.String())
				}
				heap.Free()
				m.pause()
			}
		}
	}()
	Log.Info("Manager started.")
	m.background = true
}

// Background returns true when started in a goroutine.
func (m *Manager) Background() (b bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	b = m.background
	return
}

// Iterate reaps unreferenced resources.
// Turns the 'crank' once.
// Intended to be called directly from Run() and test harnesses.
func (m *Manager) Iterate() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
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
	for _, r := range registered {
		r.Run()
	}
}

// Pause.
func (m *Manager) pause() {
	time.Sleep(Settings.Frequency.Reaper)
}

// Reaper interface.
type Reaper interface {
	Run()
}
