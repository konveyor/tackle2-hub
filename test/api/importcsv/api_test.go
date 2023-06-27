package importcsv

import (
	"strconv"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestImportCSV(t *testing.T) {
	for _, r := range TestCases {
		t.Run(r.FileName, func(t *testing.T) {

			// Upload CSV.
			inputData := api.ImportSummary{}
			assert.Must(t, Client.FilePost(api.SummariesRoot+"/upload", r.FileName, &inputData))

			// Check list of Applications.
			gotApps, _ := Application.List()
			expectedApps := r.ExpectedApplications
			if len(gotApps) != len(expectedApps) {
				t.Errorf("Mismatch in number of imported Applications: Expected %d, Actual %d", len(expectedApps), len(gotApps))
			} else {
				for i, importedApp := range gotApps {
					assert.FlatEqual(expectedApps[i].Name, importedApp.Name)
					assert.FlatEqual(expectedApps[i].Description, importedApp.Description)
					assert.FlatEqual(expectedApps[i].Repository.Kind, importedApp.Repository.Kind)
					assert.FlatEqual(expectedApps[i].Repository.URL, importedApp.Repository.URL)
					assert.FlatEqual(expectedApps[i].Binary, importedApp.Binary)
					for j, tag := range expectedApps[i].Tags {
						assert.FlatEqual(tag.Name, importedApp.Tags[j].Name)
					}
					assert.FlatEqual(expectedApps[i].BusinessService.Name, importedApp.BusinessService.Name)
				}
			}

			// Check list of Dependencies.
			gotDeps, _ := Dependency.List()
			expectedDeps := r.ExpectedDependencies
			if len(gotDeps) != len(expectedDeps) {
				t.Errorf("Mismatch in number of imported Dependencies: Expected %d, Actual %d", len(expectedDeps), len(gotDeps))
			} else {
				for i, importedDep := range gotDeps {
					expectedDep := expectedDeps[i].To.Name
					if importedDep.To.Name != expectedDep {
						t.Errorf("Mismatch in imported Dependency: Expected %s, Actual %s", expectedDep, importedDep.To.Name)
					}
				}
			}

			// fetch id of CSV file and convert it into required formats
			var inputID = strconv.FormatUint(uint64(inputData.ID), 10) // to be used for API compatibility

			var outputImportSummaries []api.ImportSummary
			outputMatchingSummary := api.ImportSummary{}
			assert.Should(t, Client.Get(api.SummariesRoot, &outputImportSummaries))
			for _, imp := range outputImportSummaries {
				if uint(imp.ID) == inputData.ID {
					outputMatchingSummary = imp
				}
			}
			assert.FlatEqual(len(expectedApps)+len(expectedDeps), outputMatchingSummary.ValidCount)

			// Get summaries of the Input ID.
			outputImportSummary := api.ImportSummary{}
			assert.Should(t, Client.Get(api.SummariesRoot+"/"+inputID, &outputImportSummary))

			// Get all imports.
			var outputImports []api.Import
			assert.Should(t, Client.Get(api.ImportsRoot, &outputImports))

			// Get import of the specific Input Id.
			outputImport := api.Import{}
			assert.Should(t, Client.Get(api.ImportsRoot+"/"+inputID, &outputImport))
		})
	}
}
