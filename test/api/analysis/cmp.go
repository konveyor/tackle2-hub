package analysis

import (
	"sort"

	"github.com/konveyor/tackle2-hub/shared/api"
)

// Eq compares two *api.Analysis.
func Eq(a, b *api.Analysis) (equal bool) {
	if a == nil || b == nil {
		return a == b
	}
	if a.Effort != b.Effort ||
		a.Commit != b.Commit ||
		a.Archived != b.Archived ||
		a.CreateUser != b.CreateUser ||
		a.UpdateUser != b.UpdateUser ||
		a.Application.ID != b.Application.ID {
		return
	}
	if len(a.Insights) != len(b.Insights) {
		return
	}
	for i := range a.Insights {
		if !EqInsight(a.Insights[i], b.Insights[i]) {
			return
		}
	}
	if len(a.Dependencies) != len(b.Dependencies) {
		return
	}
	sort.Slice(
		a.Dependencies,
		func(i, j int) bool {
			return a.Dependencies[i].Name < a.Dependencies[j].Name
		})
	sort.Slice(
		b.Dependencies,
		func(i, j int) bool {
			return b.Dependencies[i].Name < b.Dependencies[j].Name
		})
	for i := range a.Dependencies {
		if !EqDependency(a.Dependencies[i], b.Dependencies[i]) {
			return
		}
	}
	if len(a.Summary) != len(b.Summary) {
		return
	}
	for i := range a.Summary {
		if a.Summary[i] != b.Summary[i] {
			return
		}
	}
	equal = true
	return
}

// EqInsight compares two api.Insight.
func EqInsight(a, b api.Insight) (equal bool) {
	if a.RuleSet != b.RuleSet ||
		a.Rule != b.Rule ||
		a.Name != b.Name ||
		a.Description != b.Description ||
		a.Category != b.Category ||
		a.Effort != b.Effort {
		return
	}
	if len(a.Labels) != len(b.Labels) {
		return
	}
	for i := range a.Labels {
		if a.Labels[i] != b.Labels[i] {
			return
		}
	}
	if len(a.Links) != len(b.Links) {
		return
	}
	for i := range a.Links {
		if a.Links[i] != b.Links[i] {
			return
		}
	}
	if len(a.Facts) != len(b.Facts) {
		return
	}
	for k, v := range a.Facts {
		bv, ok := b.Facts[k]
		if !ok || v != bv {
			return
		}
	}
	if len(a.Incidents) != len(b.Incidents) {
		return
	}
	for i := range a.Incidents {
		if !EqIncident(a.Incidents[i], b.Incidents[i]) {
			return
		}
	}
	equal = true
	return
}

// EqIncident compares two api.Incident.
func EqIncident(a, b api.Incident) (equal bool) {
	if a.File != b.File ||
		a.Line != b.Line ||
		a.Message != b.Message ||
		a.CodeSnip != b.CodeSnip {
		return
	}
	if len(a.Facts) != len(b.Facts) {
		return
	}
	for k, v := range a.Facts {
		bv, ok := b.Facts[k]
		if !ok || v != bv {
			return
		}
	}
	equal = true
	return
}

// EqDependency compares two api.TechDependency.
func EqDependency(a, b api.TechDependency) (equal bool) {
	if a.Provider != b.Provider ||
		a.Name != b.Name ||
		a.Version != b.Version ||
		a.Indirect != b.Indirect ||
		a.SHA != b.SHA {
		return
	}
	if len(a.Labels) != len(b.Labels) {
		return
	}
	for i := range a.Labels {
		if a.Labels[i] != b.Labels[i] {
			return
		}
	}
	equal = true
	return
}
