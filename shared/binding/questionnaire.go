package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Questionnaire API.
type Questionnaire struct {
	client *Client
}

// Create a Questionnaire.
func (h *Questionnaire) Create(r *api.Questionnaire) (err error) {
	err = h.client.Post(api.QuestionnairesRoute, &r)
	return
}

// Get a Questionnaire by ID.
func (h *Questionnaire) Get(id uint) (r *api.Questionnaire, err error) {
	r = &api.Questionnaire{}
	path := Path(api.QuestionnaireRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Questionnaires.
func (h *Questionnaire) List() (list []api.Questionnaire, err error) {
	list = []api.Questionnaire{}
	err = h.client.Get(api.QuestionnairesRoute, &list)
	return
}

// Update a Questionnaire.
func (h *Questionnaire) Update(r *api.Questionnaire) (err error) {
	path := Path(api.QuestionnaireRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Questionnaire.
func (h *Questionnaire) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.QuestionnaireRoute).Inject(Params{api.ID: id}))
	return
}
