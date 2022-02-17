package k8s

import (
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//
// NewClient builds new k8s client.
func NewClient() (newClient client.Client, err error) {
	cfg, _ := config.GetConfig()
	newClient, err = client.New(
		cfg,
		client.Options{
			Scheme: scheme.Scheme,
		})
	return
}
