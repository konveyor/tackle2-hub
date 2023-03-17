package k8s

import (
	"context"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/settings"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var Settings = &settings.Settings

//
// NewClient builds new k8s client.
func NewClient() (newClient client.Client, err error) {
	if Settings.Disconnected {
		newClient = &FakeClient{}
		return
	}
	cfg, _ := config.GetConfig()
	newClient, err = client.New(
		cfg,
		client.Options{
			Scheme: scheme.Scheme,
		})
	if err != nil {
		err = liberr.Wrap(err)
	}
	return
}

type FakeClient struct {
}

func (r *FakeClient) Get(_ context.Context, _ client.ObjectKey, _ runtime.Object) (err error) {
	return
}

func (r *FakeClient) List(_ context.Context, _ *client.ListOptions, _ runtime.Object) (err error) {
	return
}

func (r *FakeClient) Create(_ context.Context, _ runtime.Object) (err error) {
	return
}

func (r *FakeClient) Delete(_ context.Context, _ runtime.Object, _ ...client.DeleteOptionFunc) (err error) {
	return
}

func (r *FakeClient) Update(_ context.Context, _ runtime.Object) (err error) {
	return
}

func (r *FakeClient) Status() (w client.StatusWriter) {
	return
}
