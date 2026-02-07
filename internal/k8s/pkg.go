package k8s

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/k8s/simulator"
	"github.com/konveyor/tackle2-hub/shared/settings"
	k8s "k8s.io/client-go/kubernetes"
	fake2 "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
		newClient = fake2.NewSimpleClientset()
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
