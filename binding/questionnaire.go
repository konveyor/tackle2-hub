package binding

import (
	"github.com/konveyor/tackle2-hub/api"
)

//
// Questionnaire API.
type Questionnaire struct {
	client *Client
}

//
// Create a Questionnaire.
func (h *Questionnaire) Create(r *api.Questionnaire) (err error) {
	err = h.client.Post(api.QuestionnairesRoot, &r)
	return
}

//
// Get a Questionnaire by ID.
func (h *Questionnaire) Get(id uint) (r *api.Questionnaire, err error) {
	r = &api.Questionnaire{}
	path := Path(api.QuestionnaireRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Questionnaires.
func (h *Questionnaire) List() (list []api.Questionnaire, err error) {
	list = []api.Questionnaire{}
	err = h.client.Get(api.QuestionnairesRoot, &list)
	return
}

//
// Update a Questionnaire.
func (h *Questionnaire) Update(r *api.Questionnaire) (err error) {
	path := Path(api.QuestionnaireRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete a Questionnaire.
func (h *Questionnaire) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.QuestionnaireRoot).Inject(Params{api.ID: id}))
	return
}
