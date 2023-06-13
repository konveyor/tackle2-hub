package importCSV

import (
	"bytes"
	"testing"
)

func TestImportCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run("CSV_Import", func(t *testing.T) {
			buffer := &bytes.Buffer{}
			err := Client.FilePost("/importsummaries/upload", r.fileName, buffer)
			if err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
