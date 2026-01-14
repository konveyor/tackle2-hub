package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/nas"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestAnalysisProfileCRUD(t *testing.T) {
	var r api.AnalysisProfile
	b, _ := json.Marshal(Base)
	_ = json.Unmarshal(b, &r)

	direct := &api.Identity{
		Name: "direct",
		Kind: "Test",
	}
	err := RichClient.Identity.Create(direct)
	assert.Must(t, err)
	defer func() {
		_ = RichClient.Identity.Delete(direct.ID)
	}()

	r.Rules.Identity = &api.Ref{ID: direct.ID, Name: direct.Name}

	err = AnalysisProfile.Create(&r)
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

func TestAnalysisProfileGetBundle(t *testing.T) {
	var r api.AnalysisProfile
	b, _ := json.Marshal(Base)
	_ = json.Unmarshal(b, &r)

	f, err := RichClient.File.Touch("Test")
	if err != nil {
		t.Fatalf(err.Error())
	}
	r.Rules.Files = append(r.Rules.Files, api.Ref{ID: f.ID})
	defer func() {
		_ = RichClient.File.Delete(f.ID)
	}()
	err = RichClient.File.Patch(f.ID, []byte("Testing."))
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = AnalysisProfile.Create(&r)
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer func() {
		_ = AnalysisProfile.Delete(r.ID)
	}()
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer func() {
		_ = nas.RmDir(tmpDir)
	}()
	err = AnalysisProfile.GetBundle(r.ID, filepath.Join(tmpDir, "bundle"))
	if err != nil {
		t.Fatalf(err.Error())
	}
}
