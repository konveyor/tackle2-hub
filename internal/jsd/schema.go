// Package jsd (json-schema definition) package provides tooling for
// managing schemas and document validation.
package jsd

import (
	"bytes"
	"encoding/json"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	yq "github.com/mikefarah/yq/v4/pkg/yqlib"
	js "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v2"
)

// Version represents a schema version.
type Version struct {
	ID         int    `json:"id"`
	Migration  string `json:"migration,omitempty" yaml:",omitempty"`
	Definition Map
}

// IsValid returns true when the jsd is valid.
func (v *Version) IsValid() (err error) {
	_, err = v.jsd()
	return
}

// Validate the specified document.
func (v *Version) Validate(document Map) (err error) {
	jsd, err := v.jsd()
	if err != nil {
		return
	}
	d := JsonSafe(document)
	err = jsd.Validate(d)
	return
}

// Migrate the specified document as needed.
func (v *Version) Migrate(document Map) (migrated Map, err error) {
	expression := v.expression()
	if expression == "" {
		migrated = document
		return
	}
	b, err := yaml.Marshal(document)
	if err != nil {
		err = &NotValid{
			Reason: err.Error(),
		}
		return
	}
	yp := yq.NewDefaultYamlPreferences()
	encoder := yq.NewYamlEncoder(yp)
	decoder := yq.NewYamlDecoder(yp)
	migrator := yq.NewStringEvaluator()
	output, err := migrator.Evaluate(
		expression,
		string(b),
		encoder,
		decoder,
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	b = []byte(output)
	migrated = Map{}
	err = yaml.Unmarshal(b, &migrated)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// Digest returns an FNV-1a digest for the version.
func (v *Version) Digest() (d string) {
	h := fnv.New64a()
	var add func(any)
	add = func(object any) {
		switch m := object.(type) {
		case Map:
			keySet := []string{}
			for k := range m {
				keySet = append(keySet, k)
			}
			sort.Strings(keySet)
			for _, k := range keySet {
				add(k)
				add(m[k])
			}
		case []Map:
			for _, v := range m {
				add(v)
			}
		default:
			b, _ := json.Marshal(m)
			_, _ = h.Write(b)
		}
	}
	add(JsonSafe(v.Definition))
	n := h.Sum64()
	d = strconv.FormatUint(n, 16)
	d = strings.ToUpper(d)
	return
}

// jsd returns a compiled json-schema object.
func (v *Version) jsd() (jsd *js.Schema, err error) {
	compiler := js.NewCompiler()
	d := JsonSafe(v.Definition)
	definition, err := json.Marshal(d)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	u := "jsd"
	err = compiler.AddResource(u, bytes.NewReader(definition))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	jsd, err = compiler.Compile(u)
	if err != nil {
		err = &NotValid{
			Reason: err.Error(),
		}
		err = liberr.Wrap(err)
	}
	return
}

// expression returns the yq migration expression.
func (v *Version) expression() (expression string) {
	expression = v.Migration
	if expression == "" {
		return
	}
	expression = strings.Replace(
		expression,
		"\n",
		"",
		-1)
	expression = strings.TrimSpace(expression)
	return
}

// Versions represents an array of schema versions.
type Versions []Version

func (v Versions) Latest() (latest Version) {
	n := len(v)
	if n > 0 {
		latest = v[n-1]
	}
	return
}

// Schema represents a document json-schema.
type Schema struct {
	Name     string   `json:"name"`
	Domain   string   `json:"domain"`
	Variant  string   `json:"variant"`
	Subject  string   `json:"subject"`
	Versions Versions `json:"versions"`
}

// With populates the object using the specified schema.
func (s *Schema) With(r *crd.Schema) {
	sp := r.Spec
	s.Domain = sp.Domain
	s.Variant = sp.Variant
	s.Subject = sp.Subject
	s.Versions = make([]Version, len(sp.Versions))
	for id := range sp.Versions {
		rv := &sp.Versions[id]
		sv := Version{
			Migration: rv.Migration,
			ID:        id,
		}
		_ = json.Unmarshal(rv.Definition.Raw, &sv.Definition)
		s.Versions[id] = sv
	}
}

// IsValid returns true when the ALL versions have valid schemas.
func (s *Schema) IsValid() (err error) {
	for _, version := range s.Versions {
		err = version.IsValid()
		if err != nil {
			return
		}
	}
	return
}

// Validate the specified document.
// Delegated to the LATEST verison.
func (s *Schema) Validate(document Map) (err error) {
	v := s.Versions.Latest()
	err = v.Validate(document)
	return
}

// Migrate the specified document as needed.
// Beginning at the `current` version +1 (index) and continuing
// through to the LATEST version.
func (s *Schema) Migrate(document Map, current int) (migrated Map, newCurrent int, err error) {
	migrated = document
	newCurrent = current
	next := current + 1
	for i := next; i < len(s.Versions); i++ {
		v := &s.Versions[i]
		migrated, err = v.Migrate(migrated)
		if err != nil {
			return
		}
		newCurrent = i
	}
	return
}

// Digest returns an FNV-1a digest for the schema.
func (s *Schema) Digest() (d string) {
	h := fnv.New64a()
	for _, v := range s.Versions {
		d := v.Digest()
		_, _ = h.Write([]byte(d))
	}
	n := h.Sum64()
	d = strconv.FormatUint(n, 16)
	d = strings.ToUpper(d)
	return
}
