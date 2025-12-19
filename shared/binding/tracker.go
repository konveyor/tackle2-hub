package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Tracker API.
type Tracker struct {
	client *Client
}

// Create a Tracker.
func (h *Tracker) Create(r *api.Tracker) (err error) {
	err = h.client.Post(api.TrackersRoute, &r)
	return
}

// Get a Tracker by ID.
func (h *Tracker) Get(id uint) (r *api.Tracker, err error) {
	r = &api.Tracker{}
	path := Path(api.TrackerRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Trackers.
func (h *Tracker) List() (list []api.Tracker, err error) {
	list = []api.Tracker{}
	err = h.client.Get(api.TrackersRoute, &list)
	return
}

// Update a Tracker.
func (h *Tracker) Update(r *api.Tracker) (err error) {
	path := Path(api.TrackerRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Tracker.
func (h *Tracker) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.TrackerRoute).Inject(Params{api.ID: id}))
	return
}

// List Projects.
func (h *Tracker) ListProjects(id uint) (projectList []api.Project, err error) {
	projectList = []api.Project{}
	err = h.client.Get(Path(api.TrackerProjectsRoute).Inject(Params{api.ID: id}), &projectList)
	return
}

// Get Projects.
func (h *Tracker) GetProjects(id1 uint, id2 uint) (project api.Project, err error) {
	project = api.Project{}
	err = h.client.Get(Path(api.TrackerProjectRoute).Inject(Params{api.ID: id1, api.ID2: id2}), &project)
	return
}

// List Project Issue Types.
func (h *Tracker) ListProjectIssueTypes(id1 uint, id2 uint) (issueType []api.IssueType, err error) {
	issueType = []api.IssueType{}
	err = h.client.Get(Path(api.TrackerProjectIssueTypesRoute).Inject(Params{api.ID: id1, api.ID2: id2}), &issueType)
	return
}
