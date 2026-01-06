package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// Tracker API.
type Tracker struct {
	client *Client
}

// Create a Tracker.
func (h *Tracker) Create(r *api2.Tracker) (err error) {
	err = h.client.Post(api2.TrackersRoute, &r)
	return
}

// Get a Tracker by ID.
func (h *Tracker) Get(id uint) (r *api2.Tracker, err error) {
	r = &api2.Tracker{}
	path := Path(api2.TrackerRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List Trackers.
func (h *Tracker) List() (list []api2.Tracker, err error) {
	list = []api2.Tracker{}
	err = h.client.Get(api2.TrackersRoute, &list)
	return
}

// Update a Tracker.
func (h *Tracker) Update(r *api2.Tracker) (err error) {
	path := Path(api2.TrackerRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a Tracker.
func (h *Tracker) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.TrackerRoute).Inject(Params{api2.ID: id}))
	return
}

// List Projects.
func (h *Tracker) ListProjects(id uint) (projectList []api2.Project, err error) {
	projectList = []api2.Project{}
	err = h.client.Get(Path(api2.TrackerProjectsRoute).Inject(Params{api2.ID: id}), &projectList)
	return
}

// Get Projects.
func (h *Tracker) GetProjects(id1 uint, id2 uint) (project api2.Project, err error) {
	project = api2.Project{}
	err = h.client.Get(Path(api2.TrackerProjectRoute).Inject(Params{api2.ID: id1, api2.ID2: id2}), &project)
	return
}

// List Project Issue Types.
func (h *Tracker) ListProjectIssueTypes(id1 uint, id2 uint) (issueType []api2.IssueType, err error) {
	issueType = []api2.IssueType{}
	err = h.client.Get(Path(api2.TrackerProjectIssueTypesRoute).Inject(Params{api2.ID: id1, api2.ID2: id2}), &issueType)
	return
}
