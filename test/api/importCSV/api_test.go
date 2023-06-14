package importCSV

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestImportCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("CSV_Import", func(t *testing.T) {

			// Upload CSV.
			buffer := make(map[string]interface{})
			err := Client.FilePost("/importsummaries/upload", r.fileName, &buffer)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get summaries.
			var destination []api.ImportSummary
			err = Client.Get("/importsummaries", &destination)
			if err != nil {
				t.Errorf(err.Error())
			}

			// check if id of output is matching with buffer id
			found := false
			for _, d := range destination {
				id := uint(buffer["id"].(float64))
				if d.ID == id {
					found = true
					break
				}
			}
			if found == false {
				t.Errorf("id of destination is matching with buffer id")
			}
		})
	}
}
