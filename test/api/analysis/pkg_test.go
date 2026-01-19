package analysis

import (
	"encoding/json"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestCreateGet(t *testing.T) {
	// Setup.
	app := &api.Application{Name: "Test"}
	err := RichClient.Application.Create(app)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Application.Delete(app.ID)
	}()

	// Create.
	var r api.Analysis
	b, _ := json.Marshal(Sample)
	_ = json.Unmarshal(b, &r)
	r.Application = api.Ref{ID: app.ID}
	err = Analysis.Create(&r)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Analysis.Delete(r.ID)
	}()

	// Fetch by id.
	got, err := Analysis.Get(r.ID)
	assert.Should(t, err)

	// Validate
	if !Eq(&r, got) {
		t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, r)
	}
}
