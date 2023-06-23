package importcsv

import (
	"strconv"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestImportCSV(t *testing.T) {
	for _, r := range TestCases {
		t.Run(r.fileName, func(t *testing.T) {

			// Upload CSV.
			inputData := make(map[string]interface{})
			err := Client.FilePost("/importsummaries/upload", r.fileName, &inputData)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check list of Applications.
			importedApps, _ := Application.List()
			expectedApps := r.ExpectedApplications
			for i, expectedApp := range expectedApps {
				if i >= len(importedApps) {
					t.Errorf("Missing imported Application: %s", expectedApp.Name)
					continue
				}
				importedApp := importedApps[i]
				if importedApp.Name != expectedApp.Name {
					t.Errorf("Mismatch in imported Application: Expected %s, Actual %s", expectedApp.Name, importedApp.Name)
				}
			}

			// Check list of Dependencies.
			importedDeps, _ := Dependency.List()
			expectedDeps := r.ExpectedDependencies
			for i, expectedDep := range expectedDeps {
				if i >= len(importedDeps) {
					t.Errorf("Missing imported Dependency: %s", expectedDep.To.Name)
					continue
				}
				importedDep := importedDeps[i].To.Name
				if importedDep != expectedDep.To.Name {
					t.Errorf("Mismatch in imported Dependency: Expected %s, Actual %s", expectedDep.To.Name, importedDep)
				}
			}

			// fetch id's
			id := uint64(inputData["id"].(float64))
			var inputID = strconv.FormatUint(id, 10)
			var output []api.ImportSummary
			err = Client.Get("/importsummaries/", &output)
			if err != nil {
				t.Errorf("Can't get summaries of all imports")
			}

			// check for the id and return valid
			// for _, imp := range output {
			// 	if uint64(imp.ID) == id {
			// 		if len(importedDeps)+len(importedApps) != imp.ValidCount {
			// 			t.Errorf("Mismatch in number of valid count")
			// 		}
			// 	}
			// }

			// Get summaries of the Input ID.
			outputImport := api.ImportSummary{}
			err = Client.Get("/importsummaries/"+inputID, &outputImport)
			if err != nil {
				t.Errorf("Could not get the CSV output")
			}

			// Delete summaries of the Input ID.
			err = Client.Delete("/importsummaries/" + inputID)
			if err != nil {
				t.Errorf("CSV delete failed")
			}
		})
	}
}
