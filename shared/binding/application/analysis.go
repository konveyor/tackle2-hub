package application

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/analysis"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// Analysis sub-resource API.
type Analysis struct {
	client *client.Client
	appId  uint
}

// Create an analysis report using the manifest at the specified path.
// The manifest contains 3 sections containing documents delimited by markers.
// The manifest must contain ALL markers even when sections are empty.
// Note: `^]` = `\x1D` = GS (group separator).
// Section markers:
//
//	^]BEGIN-MAIN^]
//	^]END-MAIN^]
//	^]BEGIN-INSIGHTS^]
//	^]END-INSIGHTS^]
//	^]BEGIN-DEPS^]
//	^]END-DEPS^]
//
// The encoding must be:
// - application/json
// - application/x-yaml
// Deprecated: Use Add().
func (h Analysis) Create(manifest, encoding string) (r *api.Analysis, err error) {
	switch encoding {
	case "":
		encoding = api.MIMEJSON
	case api.MIMEJSON,
		api.MIMEYAML:
	default:
		err = liberr.New(
			"Encoding: %s not supported",
			encoding)
	}
	r = &api.Analysis{}
	path := client.Path(api.AppAnalysesRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.FilePostEncoded(path, manifest, r, encoding)
	if err != nil {
		return
	}
	return
}

// Add an analysis report for an application.
func (h Analysis) Add(r *api.Analysis) (err error) {
	r.Application.ID = h.appId
	err = analysis.New(h.client).Create(r)
	return
}

// Get the latest analysis for an application.
func (h Analysis) Get() (r *api.Analysis, err error) {
	path := client.
		Path(api.AppAnalysisRoute).
		Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, &r)
	return
}

// GetInsights Insights analysis for an application.
func (h Analysis) GetInsights() (r []api.Insight, err error) {
	path := client.Path(api.AppAnalysisInsightsRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, &r)
	return
}

// GetDependencies dependencies for an application.
func (h Analysis) GetDependencies() (r []api.TechDependency, err error) {
	path := client.Path(api.AppAnalysisDepsRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, &r)
	return
}
