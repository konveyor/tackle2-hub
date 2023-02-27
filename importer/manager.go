package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"

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
	db := m.DB.Preload("ImportTags").Preload("ImportSummary")
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
		default:
			errMsg := ""
			if imp.RecordType1 == "" {
				errMsg = "Empty Record Type."
			} else {
				errMsg = fmt.Sprintf("Invalid or unknown Record Type '%s'. Must be '1' for Application or '2' for Dependency.", imp.RecordType1)
			}
			imp.ErrorMessage = errMsg
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
	name := strings.TrimSpace(imp.ApplicationName)
	result := m.DB.Select("id").Where("name LIKE ?", name).First(app)
	if result.Error != nil {
		imp.ErrorMessage = fmt.Sprintf("Application '%s' could not be found.", name)
		return
	}

	dep := &model.Application{}
	name = strings.TrimSpace(imp.Dependency)
	result = m.DB.Select("id").Where("name lIKE ?", name).First(dep)
	if result.Error != nil {
		imp.ErrorMessage = fmt.Sprintf("Application dependency '%s' could not be found.", name)
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

	err := dependency.Create(m.DB)
	if err != nil {
		imp.ErrorMessage = err.Error()
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
		Name:        strings.TrimSpace(imp.ApplicationName),
		Description: imp.Description,
		Comments:    imp.Comments,
	}

	if app.Name == "" {
		imp.ErrorMessage = "Application Name is mandatory."
		return
	}

	repository := api.Repository{
		Kind:   imp.RepositoryKind,
		URL:    imp.RepositoryURL,
		Branch: imp.RepositoryBranch,
		Path:   imp.RepositoryPath,
	}

	// Ensure default RepositoryType (git)
	if repository.Kind == "" {
		repository.Kind = "git"
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

	// Assign Business Service
	businessService := &model.BusinessService{}
	businessServices := []model.BusinessService{}
	m.DB.Find(&businessServices)
	normBusinessServiceName := normalizedName(imp.BusinessService)
	// Find existing BusinessService
	for _, bs := range businessServices {
		if normalizedName(bs.Name) == normBusinessServiceName {
			businessService = &bs
		}
	}
	// If not found business service in database and import specifies some non-empty business service, proceeed with create it
	if businessService.ID == 0 && normBusinessServiceName != "" {
		if imp.ImportSummary.CreateEntities {
			// Create a new BusinessService if not existed
			businessService.Name = imp.BusinessService
			result := m.DB.Create(businessService)
			if result.Error != nil {
				imp.ErrorMessage = fmt.Sprintf("BusinessService '%s' cannot be created.", imp.BusinessService)
				return
			}
		} else {
			imp.ErrorMessage = fmt.Sprintf("BusinessService '%s' could not be found.", imp.BusinessService)
			return
		}
	}
	// Assign business service to the application if was specified
	if businessService.ID != 0 {
		app.BusinessService = businessService
	}

	// Process import Tags & TagCategories
	tagCategories := []model.TagCategory{}
	m.DB.Find(&tagCategories)

	tags := []model.Tag{}
	db := m.DB.Preload("TagCategory")
	db.Find(&tags)

	for _, impTag := range imp.ImportTags {
		// Prepare normalized names for importTag
		normImpTagName := normalizedName(impTag.Name)
		normImpCategory := normalizedName(impTag.Category)

		// skip if tag name normalizes to an empty string
		if normImpTagName == "" {
			continue
		}
		// fail if the tag name is ok but the tag category normalizes to an empty string
		if normImpCategory == "" {
			imp.ErrorMessage = fmt.Sprintf("Tag '%s' has missing or invalid TagCategory.", impTag.Name)
			return
		}

		// Prepare vars for Tag and its TagCategory
		appTag := &model.Tag{}
		appTagCategory := &model.TagCategory{}

		// Find existing TagCategory
		for _, tagType := range tagCategories {
			if normalizedName(tagType.Name) == normImpCategory {
				appTagCategory = &tagType
				break
			}
		}

		// Or create TagCategory (if CreateEntities is enabled)
		if appTagCategory.ID == 0 {
			if imp.ImportSummary.CreateEntities {
				appTagCategory.Name = impTag.Category
				appTagCategory.Color = fmt.Sprintf("#%x%x%x", rand.Intn(255), rand.Intn(255), rand.Intn(255))
				result := m.DB.Create(&appTagCategory)
				if result.Error != nil {
					imp.ErrorMessage = fmt.Sprintf("TagCategory '%s' cannot be created.", impTag.Category)
					return
				}
			} else {
				imp.ErrorMessage = fmt.Sprintf("TagCategory '%s' could not be found.", impTag.Category)
				return
			}
		}
		appTag.Category = *appTagCategory

		// Find existing tag
		for _, tag := range tags {
			if normalizedName(tag.Name) == normImpTagName && normalizedName(tag.Category.Name) == normImpCategory {
				appTag = &tag
				break
			}
		}
		// Or create new tag (if CreateEntities is enabled)
		if appTag.ID == 0 {
			if imp.ImportSummary.CreateEntities {
				appTag.Name = impTag.Name
				appTag.Category = *appTagCategory
				result := m.DB.Create(&appTag)
				if result.Error != nil {
					imp.ErrorMessage = fmt.Sprintf("Tag '%s' cannot be created.", impTag.Name)
					return
				}
			} else {
				imp.ErrorMessage = fmt.Sprintf("Tag '%s' could not be found.", impTag.Name)
				return
			}
		}

		// Assign the Tag to Application's Tags
		app.Tags = append(app.Tags, *appTag)
	}

	result := m.DB.Create(app)
	if result.Error != nil {
		imp.ErrorMessage = result.Error.Error()
		return
	}

	ok = true
	return
}

//
// normalizedName transforms given name to be comparable as same with similar names
// Example: normalizedName(" F oo-123 bar! ") returns "foo123bar!"
func normalizedName(name string) (normName string) {
	invalidSymbols := regexp.MustCompile("[-_\\s]")
	normName = strings.ToLower(name)
	normName = invalidSymbols.ReplaceAllString(normName, "")
	return
}
