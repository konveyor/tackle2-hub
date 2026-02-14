package _import

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

func New(client client.RestClient) Import {
	return Import{client: client}
}

// Import API.
type Import struct {
	client client.RestClient
}

// Upload a CSV file and create an Import.
func (h Import) Upload(path string) (r *api.ImportSummary, err error) {
	r = &api.ImportSummary{}
	err = h.client.FilePost(api.UploadRoute, path, r)
	return
}

// List returns all Import records.
func (h Import) List() (list []api.Import, err error) {
	list = []api.Import{}
	err = h.client.Get(api.ImportsRoute, &list)
	return
}

func (h Import) Summary() (h2 Summary) {
	h2 = Summary{client: h.client}
	return
}

type Summary struct {
	client client.RestClient
}

// Get an Import by ID.
func (h Summary) Get(id uint) (r *api.ImportSummary, err error) {
	r = &api.ImportSummary{}
	path := client.Path(api.SummaryRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List ImportSummaries.
func (h Summary) List() (list []api.ImportSummary, err error) {
	list = []api.ImportSummary{}
	err = h.client.Get(api.SummariesRoute, &list)
	return
}

// Delete an Import.
func (h Summary) Delete(id uint) (err error) {
	err = h.client.Delete(client.Path(api.SummaryRoute).Inject(client.Params{api.ID: id}))
	return
}

// Download exports the CSV.
func (h Summary) Download(destination string) (err error) {
	err = h.client.FileGet(api.DownloadRoute, destination)
	return
}
