package api

// Document json document.
type Document struct {
	Content Map    `json:"content" binding:"required"`
	Schema  string `json:"schema,omitempty"`
}

// Schema represents a document json-schema.
type Schema struct {
	Name     string   `json:"name"`
	Domain   string   `json:"domain"`
	Variant  string   `json:"variant"`
	Subject  string   `json:"subject"`
	Versions Versions `json:"versions"`
}

// Version represents a schema version.
type Version struct {
	ID         int    `json:"id"`
	Migration  string `json:"migration,omitempty" yaml:",omitempty"`
	Definition Map
}

type Versions []Version
