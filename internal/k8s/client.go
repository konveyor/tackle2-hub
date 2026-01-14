package k8s

import (
	"context"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var Settings = &settings.Settings

// NewClient builds new k8s client.
func NewClient() (newClient client.Client, err error) {
	if Settings.Disconnected {
		newClient = &FakeClient{}
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
func NewClientSet() (newClient *k8s.Clientset, err error) {
	cfg, _ := config.GetConfig()
	cfg.QPS = 200
	cfg.Burst = 400
	cfg.UserAgent = "konveyor/hub"
	newClient, err = k8s.NewForConfig(cfg)
	err = liberr.Wrap(err)
	return
}

type FakeClient struct {
}

func (r *FakeClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) (err error) {
	return
}

func (r *FakeClient) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) (err error) {
	return
}

func (r *FakeClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) (err error) {
	return
}

func (r *FakeClient) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) (err error) {
	return
}

func (r *FakeClient) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) (err error) {
	return
}

func (r *FakeClient) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) (err error) {
	return
}

func (r *FakeClient) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) (err error) {
	return
}

func (r *FakeClient) Status() (w client.StatusWriter) {
	return
}

func (r *FakeClient) Scheme() (s *runtime.Scheme) {
	return
}

func (r *FakeClient) RESTMapper() (m meta.RESTMapper) {
	return
}
