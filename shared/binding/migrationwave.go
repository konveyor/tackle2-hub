package binding

import (
	"github.com/konveyor/tackle2-hub/shared/api"
)

// MigrationWave API.
type MigrationWave struct {
	client *Client
}

// Create a MigrationWave.
func (h MigrationWave) Create(r *api.MigrationWave) (err error) {
	err = h.client.Post(api.MigrationWavesRoute, r)
	return
}

// Get a MigrationWave by ID.
func (h MigrationWave) Get(id uint) (r *api.MigrationWave, err error) {
	r = &api.MigrationWave{}
	path := Path(api.MigrationWaveRoute).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List MigrationWaves.
func (h MigrationWave) List() (list []api.MigrationWave, err error) {
	list = []api.MigrationWave{}
	err = h.client.Get(api.MigrationWavesRoute, &list)
	return
}

// Update a MigrationWave.
func (h MigrationWave) Update(r *api.MigrationWave) (err error) {
	path := Path(api.MigrationWaveRoute).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a MigrationWave.
func (h MigrationWave) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.MigrationWaveRoute).Inject(Params{api.ID: id}))
	return
}
