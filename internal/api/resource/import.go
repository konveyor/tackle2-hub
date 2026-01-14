package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Import Statuses
const (
	InProgress = "In Progress"
	Completed  = "Completed"
)

// ImportSummary REST resource.
type ImportSummary api.ImportSummary

// With updates the resource with the model.
func (r *ImportSummary) With(m *model.ImportSummary) {
	baseWith(&r.Resource, &m.Model)
	r.Filename = m.Filename
	r.ImportTime = m.CreateTime
	r.CreateEntities = m.CreateEntities
	for _, imp := range m.Imports {
		if imp.Processed {
			if imp.IsValid {
				r.ValidCount++
			} else {
				r.InvalidCount++
			}
		}
	}
	if len(m.Imports) == r.ValidCount+r.InvalidCount {
		r.ImportStatus = Completed
	} else {
		r.ImportStatus = InProgress
	}
}
