package k8s

import (
	"context"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	fakemgr "github.com/konveyor/tackle2-hub/internal/k8s/fake"
	"github.com/konveyor/tackle2-hub/internal/k8s/simulator"
	"github.com/konveyor/tackle2-hub/shared/settings"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	Settings = &settings.Settings
	Log      = logr.New("k8s", 0)
)

// NewClient builds new k8s client.
func NewClient() (newClient client.Client, err error) {
	if Settings.Disconnected {
		newClient = simulator.New()
		return
	}
	cfg, _ := config.GetConfig()
	cfg.QPS = 200
	cfg.Burst = 400
	cfg.UserAgent = "konveyor/hub"
	newClient, err = client.New(
		cfg,
		client.Options{
			Scheme: scheme.Scheme,
		})
	err = liberr.Wrap(err)
	return
}

// NewClientSet builds new k8s client.
func NewClientSet() (newClient k8s.Interface, err error) {
	if Settings.Disconnected {
		newClient = fake.NewSimpleClientset()
		return
	}
	cfg, _ := config.GetConfig()
	cfg.QPS = 200
	cfg.Burst = 400
	cfg.UserAgent = "konveyor/hub"
	newClient, err = k8s.NewForConfig(cfg)
	err = liberr.Wrap(err)
	return
}

// Manager extends controller manager.
type Manager struct {
	manager.Manager
}

// Run starts the manager in a goroutine.
func (m *Manager) Run(ctx context.Context) {
	go func() {
		err := m.Manager.Start(ctx)
		if err != nil {
			err = liberr.Wrap(err)
			Log.Error(err, "")
			panic(err)
		}
	}()
	return
}

// NewManager builds new k8s manager.
func NewManager() (m *Manager, err error) {
	m = &Manager{}
	if Settings.Disconnected {
		m.Manager = fakemgr.NewManager(simulator.New())
		return
	}
	cfg, err := config.GetConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	m.Manager, err = manager.New(
		cfg,
		manager.Options{
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
			Cache: cache.Options{
				DefaultNamespaces: map[string]cache.Config{
					Settings.Hub.Namespace: {},
				},
			},
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
