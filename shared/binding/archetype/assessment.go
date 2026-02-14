package archetype

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Assessment sub-resource API.
type Assessment struct {
	client      client.RestClient
	archetypeId uint
}

// Create an Assessment.
func (h Assessment) Create(r *api.Assessment) (err error) {
	path := client.Path(api.ArchetypeAssessmentsRoute).Inject(client.Params{api.ID: h.archetypeId})
	err = h.client.Post(path, r)
	return
}

// List Assessments.
func (h Assessment) List() (list []api.Assessment, err error) {
	list = []api.Assessment{}
	path := client.Path(api.ArchetypeAssessmentsRoute).Inject(client.Params{api.ID: h.archetypeId})
	err = h.client.Get(path, &list)
	return
}
