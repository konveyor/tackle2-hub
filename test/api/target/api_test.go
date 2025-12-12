package target

import (
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestTargetCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			var files []*api.File
			defer func() {
				for _, f := range files {
					_ = RichClient.File.Delete(f.ID)
				}
			}()
			// Image.
			file, err := RichClient.File.Put(r.Image.Name)
			assert.Should(t, err)
			files = append(files, file)
			r.Image.ID = file.ID
			// RuleSet
			if r.RuleSet != nil {
				rules := []api.Rule{}
				for _, rule := range r.RuleSet.Rules {
					file, err := RichClient.File.Put(rule.File.Name)
					assert.Should(t, err)
					files = append(files, file)
					rules = append(rules, api.Rule{
						File: &api.Ref{
							ID: file.ID,
						},
					})
				}
				r.RuleSet.Rules = rules
			}

			// Create.
			err = Target.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Target.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			if r.RuleSet != nil {
				// Add a rule.
				file, err := RichClient.File.Put("./data/rules.yaml")
				assert.Should(t, err)
				files = append(files, file)
				r.RuleSet.Rules = append(
					r.RuleSet.Rules,
					api.Rule{
						Name: "Added",
						File: &api.Ref{
							ID: file.ID,
						},
					})
				// Rule[0] removed.
				r.RuleSet.Rules = r.RuleSet.Rules[1:]
			}
			err = Target.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}
			got, err = Target.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = Target.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			_, err = Target.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
			if r.RuleSet != nil {
				_, err = RuleSet.Get(r.RuleSet.ID)
				if err == nil {
					t.Errorf("Resource exits, but should be deleted: %v", r)
				}
			}
		})
	}
}
