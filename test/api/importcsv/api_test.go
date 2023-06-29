package importcsv

import (
	"encoding/csv"
	"os"
	"strconv"
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

			// Since uploading the CSV happens asynchronously we need to wait for the upload to check Applications and Dependencies.
			time.Sleep(time.Second * 3)

			// Check list of Applications.
			gotApps, _ := Application.List()
			expectedApps := r.ExpectedApplications
			if len(gotApps) != len(expectedApps) {
				t.Errorf("Mismatch in number of imported Applications: Expected %d, Actual %d", len(expectedApps), len(gotApps))
			} else {
				for i, importedApp := range gotApps {
					if expectedApps[i].Name != importedApp.Name {
						t.Errorf("Mismatch in name of imported Application: Expected %s, Actual %s", expectedApps[i].Name, importedApp.Name)
					}
					if expectedApps[i].Description != importedApp.Description {
						t.Errorf("Mismatch in description of imported Application: Expected %s, Actual %s", expectedApps[i].Description, importedApp.Description)
					}
					if expectedApps[i].Repository.Kind != importedApp.Repository.Kind {
						t.Errorf("Mismatch in repository's kind ofimported Application: Expected %s, Actual %s", expectedApps[i].Repository.Kind, importedApp.Repository.Kind)
					}
					if expectedApps[i].Repository.URL != importedApp.Repository.URL {
						t.Errorf("Mismatch in repository's url of imported Application: Expected %s, Actual %s", expectedApps[i].Repository.URL, importedApp.Repository.URL)
					}
					if expectedApps[i].Binary != importedApp.Binary {
						t.Errorf("Mismatch in binary of imported Application: Expected %s, Actual %s", expectedApps[i].Binary, importedApp.Binary)
					}
					for j, tag := range expectedApps[i].Tags {
						if tag.Name != importedApp.Tags[j].Name {
							t.Errorf("Mismatch in tag name of imported Application: Expected %s, Actual %s", tag.Name, importedApp.Tags[j].Name)
						}
					}
					if expectedApps[i].BusinessService.Name != importedApp.BusinessService.Name {
						t.Errorf("Mismatch in name of the BusinessService of imported Application: Expected %s, Actual %s", expectedApps[i].BusinessService.Name, importedApp.BusinessService.Name)
					}
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

			// inject import summary id into Summary root
			pathForImportSummary := binding.Path(api.SummaryRoot).Inject(binding.Params{api.ID: inputID})

			// Get summaries of the Input ID.
			outputImportSummary := api.ImportSummary{}
			assert.Should(t, Client.Get(pathForImportSummary, &outputImportSummary))

			// Get all imports.
			var outputImports []api.Import
			assert.Should(t, Client.Get(api.ImportsRoot, &outputImports))
			j, k := 0, 0
			for _, imp := range outputImports {
				if imp["recordType1"] == 1 && j < len(expectedApps) {
					// An Application with no dependencies.
					if expectedApps[j].Name != imp["applicationName"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", expectedApps[j].Name, imp["applicationName"])
					}
					if expectedApps[j].Description != imp["description"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", expectedApps[j].Description, imp["description"])
					}
					if expectedApps[j].BusinessService.Name != imp["businessService"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", expectedApps[j].BusinessService.Name, imp["businessService"])
					}
					j++
				}
				if imp["recordType1"] == 2 && k < len(expectedDeps) {
					// An Application with Dependencies.
					if expectedDeps[k].From.Name != imp["applicationName"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", expectedDeps[k].From.Name, imp["applicationName"])
					}
					if expectedDeps[k].To.Name != imp["dependency"] {
						t.Errorf("Mismatch in name of import: Expected %s, Actual %s", expectedDeps[k].To.Name, imp["dependency"])
					}
					k++
				}
			}

			// Download the csv.
			pathToOutputCSV := "downloadcsv.csv"
			err := Client.FileGet(api.DownloadRoot, pathToOutputCSV)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare contents of the csv.
			file1, err := os.Open(pathToOutputCSV)
			if err != nil {
				t.Errorf(err.Error())
				return
			}
			defer file1.Close()

			// Open the second CSV file
			file2, err := os.Open(r.FileName)
			if err != nil {
				t.Errorf(err.Error())
				return
			}
			defer file2.Close()

			// Read both the CSV files for comparison.
			reader1 := csv.NewReader(file1)
			reader2 := csv.NewReader(file2)

			column1, err := reader1.Read()
			if err != nil {
				t.Errorf(err.Error())
			}

			columm2, err := reader2.Read()
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check number of columns.
			if len(column1) != len(columm2) {
				t.Errorf("The Content of both the CSV files are different")
			}

			// Check column names.
			for i := range column1 {
				if column1[i] != columm2[i] {
					t.Errorf("Mismatch in the Column Names")
				}
			}

			// Compare rest of the contents of the files.
			reader1.FieldsPerRecord = -1
			item1, err := reader1.ReadAll()
			if err != nil {
				t.Errorf(err.Error())
			}

			reader2.FieldsPerRecord = -1
			item2, err := reader2.ReadAll()
			if err != nil {
				t.Errorf(err.Error())
			}

			// Compare number of records present
			if len(item1) != len(item2) {
				t.Errorf("Mismatch in number of records present")
			}

			// Compare each value.
			for i := 0; i < len(item1); i++ {
				if len(item1[i]) != len(item2[i]) {
					t.Errorf("Mismatch in number of values")
				}

				for j := 0; j < len(item1[i]); j++ {
					if item1[i][j] != item2[i][j] {
						t.Errorf("Mismatch in values")
					}
				}
			}

			// Delete import summary
			assert.Should(t, Client.Delete(pathForImportSummary))

			// Delete all imports
			id := 1
			for id <= len(expectedApps)+len(expectedDeps) {
				pathForImport := binding.Path(api.ImportRoot).Inject(binding.Params{api.ID: id})
				assert.Should(t, Client.Delete(pathForImport))
				id++
			}
		})
	}
}
