package importer

import (
	"context"
	"encoding/json"
	"fmt"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"strings"
	"time"
)

//
// Manager for processing application imports.
type Manager struct {
	// DB
	DB *gorm.DB
}

//
// Run the manager.
func (m *Manager) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second)
				_ = m.processImports()
			}
		}
	}()
}

//
// processImports creates applications and dependencies from
// unprocessed imports.
func (m *Manager) processImports() (err error) {
	list := []model.Import{}
	db := m.DB.Preload("ImportTags")
	result := db.Find(&list, "processed = ?", false)
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	for _, imp := range list {
		var ok bool
		switch imp.RecordType1 {
		case api.RecordTypeApplication:
			ok = m.createApplication(&imp)
		case api.RecordTypeDependency:
			ok = m.createDependency(&imp)
		}
		imp.IsValid = ok
		imp.Processed = true
		result = m.DB.Save(&imp)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}
	return
}

//
// createDependency creates an application dependency from
// a dependency import record.
func (m *Manager) createDependency(imp *model.Import) (ok bool) {
	app := &model.Application{}
	result := m.DB.Select("id").Where("name LIKE ?", imp.ApplicationName).First(app)
	if result.Error != nil {
		imp.ErrorMessage = fmt.Sprintf("Application '%s' could not be found.", imp.ApplicationName)
		return
	}

	dep := &model.Application{}
	result = m.DB.Select("id").Where("name LIKE ?", imp.Dependency).First(dep)
	if result.Error != nil {
		imp.ErrorMessage = fmt.Sprintf("Application dependency '%s' could not be found.", imp.Dependency)
		return
	}

	dependency := &model.Dependency{}
	switch strings.ToLower(imp.DependencyDirection) {
	case "northbound":
		dependency.FromID = dep.ID
		dependency.ToID = app.ID
	case "southbound":
		dependency.FromID = app.ID
		dependency.ToID = dep.ID
	}

	result = m.DB.Create(dependency)
	if result.Error != nil {
		imp.ErrorMessage = result.Error.Error()
		return
	}

	ok = true
	return
}

//
// createApplication creates an application from an
// application import record.
func (m *Manager) createApplication(imp *model.Import) (ok bool) {
	app := &model.Application{
		Name:        imp.ApplicationName,
		Description: imp.Description,
		Comments:    imp.Comments,
	}
	repository := api.Repository{
		URL:    imp.RepositoryURL,
		Branch: imp.RepositoryBranch,
		Path:   imp.RepositoryPath,
	}
	app.Repository, _ = json.Marshal(repository)

	// Validate Binary-related fields (allow all 3 empty or present)
	if imp.BinaryGroup != "" || imp.BinaryArtifact != "" || imp.BinaryVersion != "" {
		if imp.BinaryGroup == "" || imp.BinaryArtifact == "" || imp.BinaryVersion == "" {
			imp.ErrorMessage = fmt.Sprintf("Binary-related fields for application %s need to be all present or all empty", imp.ApplicationName)
			return
		}
	}

	// Build Binary attribute
	if imp.BinaryGroup != "" {
		app.Binary = fmt.Sprintf("%s:%s:%s", imp.BinaryGroup, imp.BinaryArtifact, imp.BinaryVersion)
		if imp.BinaryPackaging != "" {
			// Packaging can be empty
			app.Binary = fmt.Sprintf("%s:%s", app.Binary, imp.BinaryPackaging)
		}
	}

	businessService := &model.BusinessService{}
	result := m.DB.Select("id").Where("name LIKE ?", imp.BusinessService).First(businessService)
	if result.Error != nil {
		imp.ErrorMessage = fmt.Sprintf("BusinessService '%s' could not be found.", imp.BusinessService)
		return
	}
	app.BusinessService = businessService

	tags := []model.Tag{}
	db := m.DB.Preload("TagType")
	db.Find(&tags)
	for _, impTag := range imp.ImportTags {
		for _, tag := range tags {
			if tag.Name == impTag.Name && tag.TagType.Name == impTag.TagType {
				app.Tags = append(app.Tags, tag)
				continue
			}
		}
	}

	result = m.DB.Create(app)
	if result.Error != nil {
		imp.ErrorMessage = result.Error.Error()
		return
	}

	ok = true
	return
}
