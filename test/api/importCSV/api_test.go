package importCSV

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestImportCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("CSV_Import", func(t *testing.T) {

			// Upload CSV.
			err := Client.FilePost("/importsummaries/upload", r.fileName, &r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get summaries.
			var destination []api.ImportSummary
			err = Client.Get("/importsummaries", &destination)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
