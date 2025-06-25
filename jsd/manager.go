package jsd

import (
	"context"
	"slices"
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
	domains map[string][]Version
}

func (m *Manager) Load() (err error) {
	m.domains = make(map[string][]Version)
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
		key := m.Key(
			r.Spec.Domain,
			r.Spec.Variant,
			r.Spec.Subject)
		for n := range r.Spec.Versions {
			mv := &r.Spec.Versions[n]
			version := Version{}
			version.Id = n
			version.Migration = mv.Migration
			_ = json.Unmarshal(mv.Content.Raw, &version.Content)
			m.domains[key] =
				append(m.domains[key],
					version)
		}
	}
	return
}

func (m *Manager) Get(domain, variant, subject string) (v []Version, err error) {
	err = m.Load()
	if err != nil {
		return
	}
	key := m.Key(domain, variant, subject)
	v, found := m.domains[key]
	if !found {
		err = &NotFound{}
	}
	return
}

func (m *Manager) Latest(domain, variant, subject string) (v Version, err error) {
	err = m.Load()
	if err != nil {
		return
	}
	key := m.Key(domain, variant, subject)
	versions, found := m.domains[key]
	if found {
		slices.Reverse(versions)
		for _, version := range versions {
			v = version
			return
		}
	}
	err = &NotFound{}
	return
}

func (m *Manager) Key(domain, variant, subject string) (k string) {
	k = strings.Join(
		[]string{domain, variant, subject},
		".")
	return
}

type Version struct {
	Id        int    `json:"id"`
	Migration string `json:"migration,omitempty"`
	Content   json.Map
}
