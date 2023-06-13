package importCSV

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
)

func TestImportCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("CSV_Import", func(t *testing.T) {

			// Upload CSV.
			api.ImportHandler.UploadCSV() // need to work on this how to fetch files(use RichClient.Client)

			// Get.
			err := Client.Get(r.fileName, r.ExpectedApplications) // Client.Get() accepts a path need to figure out the path and the object
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
