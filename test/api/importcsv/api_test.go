package importcsv

import (
	"strconv"
	"testing"
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

			// Get summaries.
			id := uint64(inputData["id"].(float64))
			var inputID = strconv.FormatUint(id, 10)
			outputImport := make(map[string]interface{})
			err = Client.Get("/importsummaries/"+inputID, &outputImport)
			if err != nil {
				t.Errorf("CSV import failed")
			}
			// access the import sectio of output data to checks applications and dependencies.
			imports := outputImport["Imports"]

			// fetch ExpectedApplications and ExpectedDependencies.
			expectedApps := r.ExpectedApplications
			expectedDeps := r.ExpectedDependencies
			importList, ok := imports.([]interface{})
			if ok {
				i, j := 0, 0
				for _, item := range importList {
					if importMap, ok := item.(map[string]interface{}); ok {
						applicationName := importMap["ApplicationName"].(string)
						dependencyName := importMap["Dependency"].(string)
						// if no dependency mentioned just compare the applications.
						if len(dependencyName) == 0 {
							if applicationName != expectedApps[i].Name {
								t.Errorf("The output applications %v doesnt match with the expected applications %v", applicationName, expectedApps[i].Name)
							}
							i++
						}
						// if dependency name present, compare the names of applications and dependencies.
						if len(dependencyName) != 0 {
							if applicationName != expectedApps[i].Name || dependencyName != expectedDeps[j].To.Name {
								if applicationName != expectedApps[i].Name {
									t.Errorf("The output applications %v doesnt match with the expected applications %v", applicationName, expectedApps[i].Name)
								}
								if dependencyName != expectedDeps[j].To.Name {
									t.Errorf("The output dependency %v doesnt match with the expected dependency %v", dependencyName, expectedDeps[j].To.Name)
								}
							}
							// if there is a match increment the application by 2 as there is a dependency between 2 applications and dependency only by 1.
							i += 2
							j++
						}
					}
				}
			}

		})
	}
}
