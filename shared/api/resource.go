package api

import (
	"encoding/json"
	"time"
)

// Resource base REST resource.
type Resource struct {
	ID         uint      `json:"id,omitempty" yaml:"id,omitempty"`
	CreateUser string    `json:"createUser" yaml:"createUser,omitempty"`
	UpdateUser string    `json:"updateUser" yaml:"updateUser,omitempty"`
	CreateTime time.Time `json:"createTime" yaml:"createTime,omitempty"`
}

// Ref represents a FK.
// Contains the PK and (name) natural key.
// The name is optional and read-only.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name,omitempty"`
}

// Map unstructured object.
type Map map[string]any

// As convert the content into the object.
// The object must be a pointer.
func (m *Map) As(object any) (err error) {
	b, err := json.Marshal(m)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, object)
	return
}
