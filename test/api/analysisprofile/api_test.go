package analysisprofile

import (
	"encoding/json"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestAnalysisProfileCRUD(t *testing.T) {
	var r api.AnalysisProfile
	b, _ := json.Marshal(Base)
	_ = json.Unmarshal(b, &r)

	err := AnalysisProfile.Create(&r)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Get
	got, err := AnalysisProfile.Get(r.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !assert.Eq(got, r) {
		t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
	}

	// Update.
	r.Name = r.Name + "updated"
	err = AnalysisProfile.Update(&r)
	if err != nil {
		t.Errorf(err.Error())
	}
	got, err = AnalysisProfile.Get(r.ID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.UpdateUser = got.UpdateUser
	if !assert.Eq(got, r) {
		t.Errorf("Different response error.\nGot:\n%+v\nExpected:\n%+v", got, &r)
	}

	// Delete.
	err = AnalysisProfile.Delete(r.ID)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = AnalysisProfile.Get(r.ID)
	if err == nil {
		t.Errorf("Resource exits, but should be deleted: %v", r)
	}
}
