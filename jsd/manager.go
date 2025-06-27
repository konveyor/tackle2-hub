package jsd

import (
	"bytes"
	"context"
	"strings"

	"github.com/jortel/go-utils/logr"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/migration/json"
	"github.com/konveyor/tackle2-hub/settings"
	js "github.com/santhosh-tekuri/jsonschema/v5"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Settings = &settings.Settings
	Log      = logr.WithName("jsd-manager")
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

type Version struct {
	Name      string `json:"name"`
	Number    int    `json:"number"`
	Migration string `json:"migration,omitempty"`
	Content   json.Map
}

func (v *Version) IsValid() (err error) {
	_, err = v.jsd()
	return
}

func (v *Version) Validate(document json.Map) (err error) {
	jsd, err := v.jsd()
	if err != nil {
		return
	}
	err = jsd.Validate(document)
	return
}

func (v *Version) jsd() (jsd *js.Schema, err error) {
	compiler := js.NewCompiler()
	content, err := json.Marshal(v.Content)
	if err != nil {
		return
	}
	err = compiler.AddResource(v.Name, bytes.NewReader(content))
	if err != nil {
		return
	}
	jsd, err = compiler.Compile(v.Name)
	if err != nil {
		err = &NotValid{
			Reason: err.Error(),
		}
	}
	return
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

func (s *Schema) IsValid() (err error) {
	for _, version := range s.Versions {
		err = version.IsValid()
		if err != nil {
			return
		}
	}
	return
}

func (s *Schema) Validate(document json.Map) (err error) {
	v := s.Versions.Latest()
	err = v.Validate(document)
	return
}
