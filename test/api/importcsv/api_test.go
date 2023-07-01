package importcsv

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/binding"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestImportCSV(t *testing.T) {
	for _, r := range TestCases {
		t.Run(r.FileName, func(t *testing.T) {

			// Upload CSV.
			inputData := api.ImportSummary{}
			assert.Must(t, Client.FilePost(api.UploadRoot, r.FileName, &inputData))

			// inject import summary id into Summary root
			pathForImportSummary := binding.Path(api.SummaryRoot).Inject(binding.Params{api.ID: inputData.ID})

			// Since uploading the CSV happens asynchronously we need to wait for the upload to check Applications and Dependencies.
			time.Sleep(time.Second)

			// Code below not working as expected need to check
			
			// checkImport := api.ImportSummary{}
			// for checkImport.ValidCount + checkImport.InvalidCount != len(r.ExpectedApplications)+len(r.ExpectedDependencies){
			// 	 assert.Should(t, Client.Get(pathForImportSummary, &checkImport))
			// }

			// Check list of Applications.
			gotApps, _ := Application.List()
			if len(gotApps) != len(r.ExpectedApplications) {
				t.Errorf("Mismatch in number of imported Applications: Expected %d, Actual %d", len(r.ExpectedApplications), len(gotApps))
			} else {
				for i, importedApp := range gotApps {
					if r.ExpectedApplications[i].Name != importedApp.Name {
						t.Errorf("Mismatch in name of imported Application: Expected %s, Actual %s", r.ExpectedApplications[i].Name, importedApp.Name)
					}
					if r.ExpectedApplications[i].Description != importedApp.Description {
						t.Errorf("Mismatch in description of imported Application: Expected %s, Actual %s", r.ExpectedApplications[i].Description, importedApp.Description)
					}
					if r.ExpectedApplications[i].Repository.Kind != importedApp.Repository.Kind {
						t.Errorf("Mismatch in repository's kind ofimported Application: Expected %s, Actual %s", r.ExpectedApplications[i].Repository.Kind, importedApp.Repository.Kind)
					}
					if r.ExpectedApplications[i].Repository.URL != importedApp.Repository.URL {
						t.Errorf("Mismatch in repository's url of imported Application: Expected %s, Actual %s", r.ExpectedApplications[i].Repository.URL, importedApp.Repository.URL)
					}
					if r.ExpectedApplications[i].Binary != importedApp.Binary {
						t.Errorf("Mismatch in binary of imported Application: Expected %s, Actual %s", r.ExpectedApplications[i].Binary, importedApp.Binary)
					}
					for j, tag := range r.ExpectedApplications[i].Tags {
						if tag.Name != importedApp.Tags[j].Name {
							t.Errorf("Mismatch in tag name of imported Application: Expected %s, Actual %s", tag.Name, importedApp.Tags[j].Name)
						}
					}
					if r.ExpectedApplications[i].BusinessService.Name != importedApp.BusinessService.Name {
						t.Errorf("Mismatch in name of the BusinessService of imported Application: Expected %s, Actual %s", r.ExpectedApplications[i].BusinessService.Name, importedApp.BusinessService.Name)
					}
				}
			}

			// Check list of Dependencies.
			gotDeps, _ := Dependency.List()
			if len(gotDeps) != len(r.ExpectedDependencies) {
				t.Errorf("Mismatch in number of imported Dependencies: Expected %d, Actual %d", len(r.ExpectedDependencies), len(gotDeps))
			} else {
				for i, importedDep := range gotDeps {
					if importedDep.To.Name != r.ExpectedDependencies[i].To.Name {
						t.Errorf("Mismatch in imported Dependency: Expected %s, Actual %s", r.ExpectedDependencies[i].To.Name, importedDep.To.Name)
					}
					if importedDep.From.Name != r.ExpectedDependencies[i].From.Name {
						t.Errorf("Mismatch in imported Dependency: Expected %s, Actual %s", r.ExpectedDependencies[i].From.Name, importedDep.From.Name)
					}
				}
			}

			var outputImportSummaries []api.ImportSummary
			outputMatchingSummary := api.ImportSummary{}
			assert.Should(t, Client.Get(api.SummariesRoot, &outputImportSummaries))
			for _, imp := range outputImportSummaries {
				if uint(imp.ID) == inputData.ID {
					outputMatchingSummary = imp
				}
			}

			// Compare the number of imports.
			if len(r.ExpectedApplications)+len(r.ExpectedDependencies) != outputMatchingSummary.ValidCount {
				t.Errorf("Mismatch in no. of imports: expected %v, got %v", len(r.ExpectedApplications)+len(r.ExpectedDependencies), outputMatchingSummary.ValidCount)
			}

			// Get summaries of the Input ID.
			outputImportSummary := api.ImportSummary{}
			assert.Should(t, Client.Get(pathForImportSummary, &outputImportSummary))

			// Get all imports.
			var outputImports []api.Import
			assert.Should(t, Client.Get(api.ImportsRoot, &outputImports))
			
			// Check for number of imports.
			if len(outputImports) != len(r.ExpectedApplications)+len(r.ExpectedDependencies) {
				t.Errorf("Mismatch in number of imports")
			}
			
			// Checks for individual applications and dependencies.
			j, k := 0, 0
			for _, imp := range outputImports {
				if imp["recordType1"] == 1 && j < len(r.ExpectedApplications) {
					// An Application with no dependencies.
					if r.ExpectedApplications[j].Name != imp["applicationName"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", r.ExpectedApplications[j].Name, imp["applicationName"])
					}
					if r.ExpectedApplications[j].Description != imp["description"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", r.ExpectedApplications[j].Description, imp["description"])
					}
					if r.ExpectedApplications[j].BusinessService.Name != imp["businessService"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", r.ExpectedApplications[j].BusinessService.Name, imp["businessService"])
					}
					j++
				}
				if imp["recordType1"] == 2 && k < len(r.ExpectedDependencies) {
					// An Application with Dependencies.
					if r.ExpectedDependencies[k].From.Name != imp["applicationName"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", r.ExpectedDependencies[k].From.Name, imp["applicationName"])
					}
					if r.ExpectedDependencies[k].To.Name != imp["dependency"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", r.ExpectedDependencies[k].To.Name, imp["dependency"])
					}
					k++
				}
			}

			// Download the csv.
			pathToGotCSV := "downloadcsv.csv"
			assert.Should(t, Client.FileGet(api.DownloadRoot, pathToGotCSV))

			// Read the got CSV file.
			gotCSV, err := ioutil.ReadFile(pathToGotCSV)
			if err != nil {
				t.Errorf("Error reading CSV: %s", pathToGotCSV)
			}
			gotCSVString := string(gotCSV)

			// Read the expected CSV file.
			expectedCSV, err := ioutil.ReadFile(r.FileName)
			if err != nil {
				t.Errorf("Error reading CSV: %s", r.FileName)
			}
			expectedCSVString := string(expectedCSV)
			if gotCSVString != expectedCSVString {
				t.Errorf("The CSV files have different content %s and %s", gotCSVString, expectedCSVString)
			}

			// Remove the CSV file created.
			err = os.Remove(pathToGotCSV)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Delete imported summaries
			assert.Must(t, Client.Delete(pathForImportSummary))

			// Delete imported Applications.
			for _, apps := range gotApps {
				assert.Must(t, Application.Delete(apps.ID))
			}

			// Delete imported Dependencies.
			for _, deps := range gotDeps {
				assert.Must(t, Dependency.Delete(deps.ID))
			}
		})
	}
}
