package jsd

import (
	"context"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Settings = &settings.Settings
	Log      = logr.New("jsd-manager", 0)
)

// Manager maintains the schema inventory.
type Manager struct {
	Client    client.Client
	domains   map[string]Schema
	names     map[string]Schema
	hasLoaded bool
}

// Load inventory.
func (m *Manager) Load() (err error) {
	if m.hasLoaded {
		return
	}
	m.domains = make(map[string]Schema)
	m.names = make(map[string]Schema)
	if Settings.Disconnected {
		return
	}
	list := &crd.SchemaList{}
	err = m.Client.List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Namespace,
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		schema := Schema{Name: r.Name}
		schema.With(r)
		if !m.Validate(&schema) {
			continue
		}
		m.names[r.Name] = schema
		key := m.Key(
			r.Spec.Domain,
			r.Spec.Variant,
			r.Spec.Subject)
		m.domains[key] = schema
	}
	m.hasLoaded = true
	return
}

// Get schema by name.
func (m *Manager) Get(name string) (schema Schema, err error) {
	err = m.Load()
	if err != nil {
		return
	}
	schema, found := m.names[name]
	if !found {
		err = &NotFound{}
	}
	return
}

// List schemas.
func (m *Manager) List() (list []Schema, err error) {
	err = m.Load()
	if err != nil {
		return
	}
	for _, s := range m.names {
		list = append(list, s)
	}
	return
}

// Find schema by name, variant, and subject.
func (m *Manager) Find(domain, variant, subject string) (s Schema, err error) {
	err = m.Load()
	if err != nil {
		return
	}
	key := m.Key(domain, variant, subject)
	s, found := m.domains[key]
	if !found {
		err = &NotFound{}
	}
	return
}

// Key returns a key based on name, variant, and subject.
func (m *Manager) Key(domain, variant, subject string) (k string) {
	k = strings.Join(
		[]string{domain, variant, subject},
		".")
	return
}

// Validate the schema.
func (m *Manager) Validate(schema *Schema) (valid bool) {
	err := schema.IsValid()
	if err == nil {
		valid = true
		return
	}
	Log.Error(
		err,
		"Schema not valid.",
		"name",
		schema.Name)
	return
}
