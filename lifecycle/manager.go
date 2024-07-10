package lifecycle

import (
	"sync"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/database"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

var Log = logr.WithName("lifecycle")

type Runnable interface {
	Run(ctx context.Context, wg *sync.WaitGroup)
}

func NewManager(db *gorm.DB) (m *Manager) {
	m = &Manager{
		db: db,
	}
	return
}

type Manager struct {
	db       *gorm.DB
	running  bool
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	runnable []Runnable
}

func (r *Manager) Register(runnable Runnable) {
	r.runnable = append(r.runnable, runnable)
}

func (r *Manager) Run() {
	if r.running {
		return
	}
	r.wg = sync.WaitGroup{}
	var ctx context.Context
	ctx, r.cancel = context.WithCancel(context.Background())
	for _, runnable := range r.runnable {
		r.wg.Add(1)
		runnable.Run(ctx, &r.wg)
	}
	r.running = true
}

func (r *Manager) Stop() {
	if !r.running {
		return
	}
	r.cancel()
	r.wg.Wait()
	err := database.Close(r.db)
	if err != nil {
		Log.Error(err, "could not close database after shutting down")
		panic(err)
	}
	r.running = false
}
