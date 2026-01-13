package binding

import (
	api2 "github.com/konveyor/tackle2-hub/shared/api"
)

// Questionnaire API.
type Questionnaire struct {
	client *Client
}

// Create a Questionnaire.
func (h *Questionnaire) Create(r *api2.Questionnaire) (err error) {
	err = h.client.Post(api2.QuestionnairesRoute, &r)
	return
}

// Get a Questionnaire by ID.
func (h *Questionnaire) Get(id uint) (r *api2.Questionnaire, err error) {
	r = &api2.Questionnaire{}
	path := Path(api2.QuestionnaireRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Questionnaires.
func (h *Questionnaire) List() (list []api2.Questionnaire, err error) {
	list = []api2.Questionnaire{}
	err = h.client.Get(api2.QuestionnairesRoute, &list)
	return
}

// Update a Questionnaire.
func (h *Questionnaire) Update(r *api2.Questionnaire) (err error) {
	path := Path(api2.QuestionnaireRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Questionnaire.
func (h *Questionnaire) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.QuestionnaireRoute).Inject(Params{api2.ID: id}))
	return
}
