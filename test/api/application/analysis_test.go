package application

import (
	"encoding/json"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	at "github.com/konveyor/tackle2-hub/test/api/analysis"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestApplicationAnalysis(t *testing.T) {
	// Setup.
	app := &api.Application{
		Name:        "Test App for Analysis",
		Description: "Application for testing analysis",
	}
	err := RichClient.Application.Create(app)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Application.Delete(app.ID)
	}()

	// Create.
	var r api.Analysis
	b, _ := json.Marshal(at.Sample)
	_ = json.Unmarshal(b, &r)
	r.Application = api.Ref{ID: app.ID}
	err = RichClient.Analysis.Create(&r)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Analysis.Delete(r.ID)
	}()

	// Get the analysis back.
	analysis := Application.Analysis(app.ID)
	got, err := analysis.Get()
	assert.Must(t, err)
	if !at.Eq(&r, got) {
		t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, r)
	}

	// Test list insights.
	gotInsights, err := analysis.ListInsights()
	if len(r.Insights) != len(gotInsights) {
		return
	}
	for i := range r.Insights {
		if !at.EqInsight(r.Insights[i], gotInsights[i]) {
			return
		}
	}

	// Test list insights.
	gotDeps, err := analysis.ListDependencies()
	if len(r.Dependencies) != len(gotDeps) {
		return
	}
	for i := range r.Dependencies {
		if !at.EqDependency(r.Dependencies[i], gotDeps[i]) {
			return
		}
	}
}
