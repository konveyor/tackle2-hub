package ruleset

import (
	"testing"

	"github.com/konveyor/tackle2-hub/api"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestRuleSetCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Prepare rules files.
			ruleFiles := []api.File{}
			rules := []api.Rule{}
			for _, rule := range r.Rules {
				ruleFile, err := RichClient.File.Put(rule.File.Name)
				assert.Should(t, err)
				rules = append(rules, api.Rule{
					File: &api.Ref{
						ID: ruleFile.ID,
					},
				})
				ruleFiles = append(ruleFiles, *ruleFile)
			}
			r.Rules = rules

			// Create.
			err := RuleSet.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := RuleSet.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			err = RuleSet.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = RuleSet.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}

			// Delete.
			err = RuleSet.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = RuleSet.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}

			// Delete rules file.
			for _, ruleFile := range ruleFiles {
				assert.Must(t, RichClient.File.Delete(ruleFile.ID))
			}
		})
	}
}
