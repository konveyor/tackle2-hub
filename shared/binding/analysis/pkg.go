package analysis

import (
	"os"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
	"gopkg.in/yaml.v3"
)

func New(client client.RestClient) (h Analysis) {
	h = Analysis{client: client}
	return
}

// Analysis API.
type Analysis struct {
	client client.RestClient
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

// List Analyses.
func (h Analysis) List() (r []api.Analysis, err error) {
	r = []api.Analysis{}
	err = h.client.Get(api.AnalysesRoute, &r)
	return
}

// Delete an Analysis by ID.
func (h Analysis) Delete(id uint) (err error) {
	path := client.Path(api.AnalysisRoute).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}

// Archive an Analysis by ID.
func (h Analysis) Archive(id uint) (err error) {
	path := client.Path(api.AnalysisArchiveRoute).Inject(client.Params{api.ID: id})
	err = h.client.Post(path, nil)
	return
}

// ListDependencies returns global list of dependencies across all analyses.
func (h Analysis) ListDependencies() (list []api.TechDependency, err error) {
	list = []api.TechDependency{}
	err = h.client.Get(api.AnalysesDepsRoute, &list)
	return
}

// ListInsights returns global list of insights across all analyses.
func (h Analysis) ListInsights() (list []api.Insight, err error) {
	list = []api.Insight{}
	err = h.client.Get(api.AnalysesInsightsRoute, &list)
	return
}

// GetInsight retrieves a specific insight by ID.
func (h Analysis) GetInsight(id uint) (r *api.Insight, err error) {
	r = &api.Insight{}
	path := client.Path(api.AnalysesInsightRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// ListIncidents returns global list of incidents across all analyses.
func (h Analysis) ListIncidents() (list []api.Incident, err error) {
	list = []api.Incident{}
	err = h.client.Get(api.AnalysesIncidentsRoute, &list)
	return
}

// GetIncident retrieves a specific incident by ID.
func (h Analysis) GetIncident(id uint) (r *api.Incident, err error) {
	r = &api.Incident{}
	path := client.Path(api.AnalysesIncidentRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// Select returns the API for a selected analysis.
func (h Analysis) Select(id uint) (h2 Selected) {
	h2 = Selected{
		client:     h.client,
		analysisId: id,
	}
	return
}

// Selected analysis API.
type Selected struct {
	client     client.RestClient
	analysisId uint
}

// ListInsights returns insights for the selected analysis.
func (h Selected) ListInsights() (list []api.Insight, err error) {
	list = []api.Insight{}
	path := client.Path(api.AnalysisInsightsRoute).Inject(client.Params{api.ID: h.analysisId})
	err = h.client.Get(path, &list)
	return
}

// GetInsight retrieves a specific insight and its incidents.
func (h Selected) GetInsight(id uint) (h2 SelectedInsight) {
	h2 = SelectedInsight{
		client:     h.client,
		analysisId: h.analysisId,
		insightId:  id,
	}
	return
}

// SelectedInsight API for a specific insight within an analysis.
type SelectedInsight struct {
	client     client.RestClient
	analysisId uint
	insightId  uint
}

// ListIncidents returns incidents for the selected insight.
func (h SelectedInsight) ListIncidents() (list []api.Incident, err error) {
	list = []api.Incident{}
	path := client.Path(api.AnalysisIncidentsRoute).
		Inject(client.Params{api.ID: h.insightId})
	err = h.client.Get(path, &list)
	return
}
