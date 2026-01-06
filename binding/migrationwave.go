package binding

import (
	api2 "github.com/konveyor/tackle2-hub/api"
)

// MigrationWave API.
type MigrationWave struct {
	client *Client
}

// Create a MigrationWave.
func (h *MigrationWave) Create(r *api2.MigrationWave) (err error) {
	err = h.client.Post(api2.MigrationWavesRoute, &r)
	return
}

// Get a MigrationWave by ID.
func (h *MigrationWave) Get(id uint) (r *api2.MigrationWave, err error) {
	r = &api2.MigrationWave{}
	path := Path(api2.MigrationWaveRoute).Inject(Params{api2.ID: id})
	err = h.client.Get(path, r)
	return
}

// List MigrationWaves.
func (h *MigrationWave) List() (list []api2.MigrationWave, err error) {
	list = []api2.MigrationWave{}
	err = h.client.Get(api2.MigrationWavesRoute, &list)
	return
}

// Update a MigrationWave.
func (h *MigrationWave) Update(r *api2.MigrationWave) (err error) {
	path := Path(api2.MigrationWaveRoute).Inject(Params{api2.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a MigrationWave.
func (h *MigrationWave) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api2.MigrationWaveRoute).Inject(Params{api2.ID: id}))
	return
}
