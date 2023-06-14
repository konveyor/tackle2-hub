package importCSV

import (
	"strconv"
	"testing"
)

func TestImportCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("CSV_Import", func(t *testing.T) {

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
		})
	}
}
