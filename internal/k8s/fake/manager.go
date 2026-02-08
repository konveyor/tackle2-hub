package fake

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1alpha1 "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Manager is a no-op manager for testing/disconnected mode.
// It implements the manager.Manager interface but does nothing.
type Manager struct {
	client client.Client
}

// NewManager creates a new fake manager with the given client.
func NewManager(client client.Client) *Manager {
	return &Manager{
		client: client,
	}
}

// GetClient returns the client.
func (m *Manager) GetClient() (c client.Client) {
	c = m.client
	return
}

// GetAPIReader returns the client (same as GetClient for fake).
func (m *Manager) GetAPIReader() (r client.Reader) {
	r = m.client
	return
}

// Start blocks until the context is canceled (like a real manager).
func (m *Manager) Start(ctx context.Context) (err error) {
	<-ctx.Done()
	return
}

// No-op implementations for other manager.Manager interface methods

func (m *Manager) GetConfig() (cfg *rest.Config) {
	return
}

func (m *Manager) GetScheme() (s *runtime.Scheme) {
	return
}

func (m *Manager) GetRESTMapper() (mapper meta.RESTMapper) {
	return
}

func (m *Manager) GetCache() (c cache.Cache) {
	return
}

func (m *Manager) GetEventRecorderFor(name string) (recorder record.EventRecorder) {
	return
}

func (m *Manager) GetFieldIndexer() (indexer client.FieldIndexer) {
	return
}

func (m *Manager) Add(manager.Runnable) (err error) {
	return
}

func (m *Manager) Elected() (ch <-chan struct{}) {
	return
}

func (m *Manager) AddMetricsExtraHandler(path string, handler http.Handler) (err error) {
	return
}

func (m *Manager) AddHealthzCheck(name string, check healthz.Checker) (err error) {
	return
}

func (m *Manager) AddReadyzCheck(name string, check healthz.Checker) (err error) {
	return
}

func (m *Manager) GetWebhookServer() (server *webhook.Server) {
	return
}

func (m *Manager) GetLogger() (log logr.Logger) {
	log = logr.Discard()
	return
}

func (m *Manager) GetControllerOptions() (spec v1alpha1.ControllerConfigurationSpec) {
	return
}

func (m *Manager) SetFields(interface{}) (err error) {
	return
}
