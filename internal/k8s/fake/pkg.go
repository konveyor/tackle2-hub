package fake

import (
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
}

func (r *Client) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) (err error) {
	return
}

func (r *Client) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) (err error) {
	return
}

func (r *Client) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) (err error) {
	return
}

func (r *Client) Delete(_ context.Context, _ client.Object, _ ...client.DeleteOption) (err error) {
	return
}

func (r *Client) DeleteAllOf(_ context.Context, _ client.Object, _ ...client.DeleteAllOfOption) (err error) {
	return
}

func (r *Client) Update(_ context.Context, _ client.Object, _ ...client.UpdateOption) (err error) {
	return
}

func (r *Client) Patch(_ context.Context, _ client.Object, _ client.Patch, _ ...client.PatchOption) (err error) {
	return
}

func (r *Client) Status() (w client.StatusWriter) {
	return
}

func (r *Client) Scheme() (s *runtime.Scheme) {
	return
}

func (r *Client) RESTMapper() (m meta.RESTMapper) {
	return
}
