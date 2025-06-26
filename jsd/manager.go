package jsd

import (
	"context"
	"strings"

	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/migration/json"
	"github.com/konveyor/tackle2-hub/settings"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Settings = &settings.Settings
)

type Manager struct {
	Client  client.Client
	domains map[string]Schema
	names   map[string]Schema
}

func (m *Manager) Load() (err error) {
	m.domains = make(map[string]Schema)
	m.names = make(map[string]Schema)
	list := &crd.SchemaList{}
	err = m.Client.List(
		context.TODO(),
		list,
		&k8s.ListOptions{
			Namespace: Settings.Namespace,
		})
	if err != nil {
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		schema := Schema{Name: r.Name}
		schema.With(r)
		m.names[r.Name] = schema
		key := m.Key(
			r.Spec.Domain,
			r.Spec.Variant,
			r.Spec.Subject)
		m.domains[key] = schema
	}
	return
}

func (m *Manager) Get(name string) (s Schema, err error) {
	err = m.Load()
	if err != nil {
		return
	}
	s, found := m.names[name]
	if !found {
		err = &NotFound{}
	}
	return
}

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

func (m *Manager) Key(domain, variant, subject string) (k string) {
	k = strings.Join(
		[]string{domain, variant, subject},
		".")
	return
}

type Version struct {
	Name      string `json:"name"`
	Number    int    `json:"number"`
	Migration string `json:"migration,omitempty"`
	Content   json.Map
}

type Versions []Version

func (v Versions) Latest() (latest Version) {
	n := len(v)
	if n > 0 {
		latest = v[n-1]
	}
	return
}

type Schema struct {
	Name     string   `json:"name"`
	Domain   string   `json:"domain,omitempty"`
	Variant  string   `json:"variant,omitempty"`
	Subject  string   `json:"subject,omitempty"`
	Versions Versions `json:"versions,omitempty"`
}

func (s *Schema) With(r *crd.Schema) {
	sp := r.Spec
	s.Domain = sp.Domain
	s.Variant = sp.Variant
	s.Subject = sp.Subject
	s.Versions = make([]Version, len(sp.Versions))
	for i := range sp.Versions {
		rv := &sp.Versions[i]
		sv := Version{Migration: rv.Migration}
		_ = json.Unmarshal(rv.Content.Raw, &sv.Content)
		s.Versions[i] = sv
	}
}
