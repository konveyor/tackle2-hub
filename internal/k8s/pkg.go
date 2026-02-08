package k8s

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/controller"
	fakemgr "github.com/konveyor/tackle2-hub/internal/k8s/fake"
	"github.com/konveyor/tackle2-hub/internal/k8s/simulator"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gorm.io/gorm"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var Settings = &settings.Settings

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

// NewManager builds new k8s manager.
// In disconnected mode, returns a fake manager that does nothing.
// Otherwise, creates a real manager with addon controller.
func NewManager(db *gorm.DB) (mgr manager.Manager, err error) {
	if Settings.Disconnected {
		client := simulator.New()
		mgr = fakemgr.NewManager(client)
		return
	}
	cfg, err := config.GetConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	mgr, err = manager.New(
		cfg,
		manager.Options{
			MetricsBindAddress: "0",
			Namespace:          Settings.Hub.Namespace,
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = controller.Add(mgr, db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
