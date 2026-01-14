package json

import (
	"github.com/konveyor/tackle2-hub/internal/jsd"
	"gopkg.in/yaml.v2"
)

// Ref represents a FK.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name,omitempty" yaml:",omitempty"`
}

// Map alias.
type Map = map[string]any

// Any alias.
type Any any

// Data json any field.
type Data struct {
	Any
}

// Merge merges the other into self.
// Both must be a map.
func (d *Data) Merge(other Data) (merged bool) {
	b, isMap := d.AsMap()
	if !isMap {
		return
	}
	a, isMap := other.AsMap()
	if !isMap {
		return
	}
	d.Any = d.merge(a, b)
	d.Any = jsd.JsonSafe(d.Any)
	merged = true
	return
}

// Merge maps B into A.
// The B map takes precedence.
func (d *Data) merge(a, b map[any]any) (out map[any]any) {
	if a == nil {
		a = make(map[any]any)
	}
	if b == nil {
		b = make(map[any]any)
	}
	out = make(map[any]any)
	for k, v := range a {
		out[k] = v
		if bv, found := b[k]; found {
			out[k] = bv
			if av, cast := v.(map[any]any); cast {
				if bv, cast := bv.(map[any]any); cast {
					out[k] = d.merge(av, bv)
				} else {
					out[k] = bv
				}
			}
		}
	}
	for k, v := range b {
		if _, found := a[k]; !found {
			out[k] = v
		}
	}

	return
}

// AsMap returns self as a map.
func (d *Data) AsMap() (mp map[any]any, isMap bool) {
	if d.Any == nil {
		return
	}
	b, err := yaml.Marshal(d.Any)
	if err != nil {
		return
	}
	mp = make(map[any]any)
	err = yaml.Unmarshal(b, &mp)
	if err != nil {
		return
	}
	isMap = true
	return
}

// Document json document with (json-schema).
type Document struct {
	Content Map    `json:"content" binding:"required"`
	Schema  string `json:"schema,omitempty"`
}

// Validate the content with the schema.
func (d *Document) Validate(m *jsd.Manager) (err error) {
	schema, err := m.Get(d.Schema)
	if err != nil {
		return
	}
	err = schema.Validate(d.Content)
	return
}
