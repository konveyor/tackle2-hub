package analysis

import (
	"os"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
	"gopkg.in/yaml.v3"
)

func New(client *client.Client) (h Analysis) {
	h = Analysis{client: client}
	return
}

// Analysis API.
type Analysis struct {
	client *client.Client
}

// Create an Analysis.
func (h Analysis) Create(r *api.Analysis) (err error) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		return
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	_, _ = file.Write([]byte(api.BeginMainMarker))
	_, _ = file.Write([]byte{'\n'})
	main := &api.Analysis{}
	main.Application.ID = r.ID
	main.Effort = r.Effort
	main.Commit = r.Commit
	encoder := yaml.NewEncoder(file)
	err = encoder.Encode(main)
	if err != nil {
		return
	}
	_, _ = file.Write([]byte(api.EndMainMarker))
	_, _ = file.Write([]byte{'\n'})
	_, _ = file.Write([]byte(api.BeginInsightsMarker))
	_, _ = file.Write([]byte{'\n'})
	for _, insight := range r.Insights {
		err = encoder.Encode(insight)
		if err != nil {
			return
		}
	}
	_, _ = file.Write([]byte(api.EndInsightsMarker))
	_, _ = file.Write([]byte{'\n'})
	_, _ = file.Write([]byte(api.BeginDepsMarker))
	_, _ = file.Write([]byte{'\n'})
	for _, dep := range r.Dependencies {
		err = encoder.Encode(dep)
		if err != nil {
			return
		}
	}
	_, _ = file.Write([]byte(api.EndDepsMarker))
	err = file.Close()
	if err != nil {
		return
	}
	path := client.Path(api.AppAnalysesRoute).Inject(client.Params{api.ID: r.Application.ID})
	err = h.client.FilePostEncoded(path, file.Name(), r, api.MIMEYAML)
	return
}

// Get an Analysis by ID.
func (h Analysis) Get(id uint) (r *api.Analysis, err error) {
	r = &api.Analysis{}
	path := client.Path(api.AnalysisRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// Delete an Analysis by ID.
func (h Analysis) Delete(id uint) (err error) {
	path := client.Path(api.AnalysisRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}
