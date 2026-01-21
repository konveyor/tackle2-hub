package application

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/bucket"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Bucket returns the bucket API.
// Deprecated: Use Application.Select(id).Analysis instead.
func (h Application) Bucket(id uint) (h2 bucket.Content) {
	selected := h.Select(id)
	h2 = selected.Bucket
	return
}

// Tags returns the tags API.
// Deprecated: Use Application.Select(id).Analysis instead.
func (h Application) Tags(id uint) (h2 TagXX) {
	h2 = TagXX{
		client: h.client,
		appId:  id,
	}
	return
}

// Facts returns the facts API.
// Deprecated: Use Application.Select(id).Analysis instead.
func (h Application) Facts(id uint) (h2 FactXX) {
	h2 = FactXX{
		client: h.client,
		appId:  id,
	}
	return
}

// Analysis returns the analysis API.
// Deprecated: Use Application.Select(id).Analysis instead.
func (h Application) Analysis(id uint) (h2 AnalysisXX) {
	h2 = AnalysisXX{
		client: h.client,
		appId:  id,
	}
	return
}

// Manifest returns the manifest API.
// Deprecated: Use Application.Select(id).Manifest instead.
func (h Application) Manifest(id uint) (h2 Manifest) {
	selected := h.Select(id)
	h2 = selected.Manifest
	return
}

// Identity returns the identity API.
// Deprecated: Use Application.Select(id).Identity instead.
func (h Application) Identity(id uint) (h2 Identity) {
	selected := h.Select(id)
	h2 = selected.Identity
	return
}

// Assessment returns the assessment API.
// Deprecated: Use Application.Select(id).Assessment instead.
func (h Application) Assessment(id uint) (f Assessment) {
	selected := h.Select(id)
	f = selected.Assessment
	return
}

// AnalysisXX analysis API.
// Deprecated: Use Analysis instead.
type AnalysisXX struct {
	client client.RestClient
	appId  uint
}

// Create an analysis.
func (h AnalysisXX) Create(manifest, encoding string) (r *api.Analysis, err error) {
	h2 := Analysis{
		client: h.client,
		appId:  h.appId,
	}
	return h2.Upload(manifest, encoding)
}

// TagXX sub-resource API.
// Provides association management of tags to applications by name.
// Deprecated: Use Tag instead.
type TagXX struct {
	client client.RestClient
	appId  uint
	source *string
}

// Source sets the source for other operations on the associated tags.
func (h *TagXX) Source(name string) {
	h.source = &name
}

// Replace the associated tags for the source with a new set.
// Returns an error if the source is not set.
func (h *TagXX) Replace(ids []uint) (err error) {
	h2 := Tag{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Replace(ids)
}

// List associated tags.
// Returns a list of tag names.
func (h *TagXX) List() (list []api.TagRef, err error) {
	h2 := Tag{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.List()
}

// Add associates a tag with the application.
func (h *TagXX) Add(id uint) (err error) {
	h2 := Tag{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Add(id)
}

// Ensure ensures tag is associated with the application.
func (h *TagXX) Ensure(id uint) (err error) {
	h2 := Tag{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Ensure(id)
}

// Delete ensures the tag is not associated with the application.
func (h *TagXX) Delete(id uint) (err error) {
	h2 := Tag{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Delete(id)
}

// FactXX sub-resource API.
// Provides association management of facts.
// Deprecated: Use Fact instead.
type FactXX struct {
	client client.RestClient
	appId  uint
	source string
}

// Source sets the source for other operations on the facts.
func (h *FactXX) Source(source string) {
	h.source = source
}

// List facts.
func (h *FactXX) List() (facts api.Map, err error) {
	h2 := Fact{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.List()
}

// Get a fact.
func (h *FactXX) Get(name string, value any) (err error) {
	h2 := Fact{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Get(name, value)
}

// Set a fact (created as needed).
func (h *FactXX) Set(name string, value any) (err error) {
	h2 := Fact{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Set(name, value)
}

// Delete a fact.
func (h *FactXX) Delete(name string) (err error) {
	h2 := Fact{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Delete(name)
}

// Replace facts.
func (h *FactXX) Replace(facts api.Map) (err error) {
	h2 := Fact{
		client: h.client,
		appId:  h.appId,
		source: h.source,
	}
	return h2.Replace(facts)
}
