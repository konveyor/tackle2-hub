package importcsv

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestImportCSV(t *testing.T) {
	for _, r := range TestCases {
		t.Run(r.FileName, func(t *testing.T) {

			// Upload CSV.
			inputData := make(map[string]interface{})
			err := Client.FilePost("/importsummaries/upload", r.FileName, &inputData)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Check list of Applications.
			importedApps, _ := Application.List()
			expectedApps := r.ExpectedApplications
			for i, expectedApp := range expectedApps {
				importedApp := importedApps[i]
				if importedApp.Name != expectedApp.Name {
					t.Errorf("Mismatch in imported Application: Expected %s, Actual %s", expectedApp.Name, importedApp.Name)
				}
			}

			// Check list of Dependencies.
			importedDeps, _ := Dependency.List()
			expectedDeps := r.ExpectedDependencies
			for i, expectedDep := range expectedDeps {
				importedDep := importedDeps[i].To.Name
				if importedDep != expectedDep.To.Name {
					t.Errorf("Mismatch in imported Dependency: Expected %s, Actual %s", expectedDep.To.Name, importedDep)
				}
			}

			// fetch id of CSV file and convert it into required formats
			id := uint64(inputData["id"].(float64))
			var inputID = strconv.FormatUint(id, 10) // to be used for API compatibility

			var outputImportSummaries []api.ImportSummary
			outputMatchingSummary := api.ImportSummary{}
			err = Client.Get("/importsummaries", &outputImportSummaries)
			if err != nil {
				t.Errorf("failed to get import summary: %v", err)
			}
			for _, imp := range outputImportSummaries {
				if uint64(imp.ID) == id {
					outputMatchingSummary = imp
				}
			}
			fmt.Println(outputMatchingSummary)
			if len(importedDeps)+len(importedApps) != outputMatchingSummary.ValidCount {
				t.Errorf("valid count not matching with number of applications and dependencies")
			}

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

			// Delete related Applications.
			err = Application.Delete(uint(id))
			if err != nil {
				t.Errorf("Application delete failed")
			}

			// Delete related Dependencies.
			err = Dependency.Delete(uint(id))
			if err != nil {
				t.Errorf("Dependency delete failed")
			}
		})
	}
}
