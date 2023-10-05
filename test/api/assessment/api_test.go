package assessment

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestAssessmentCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(fmt.Sprintf("%s for application %s", r.Questionnaire.Name, r.Application.Name), func(t *testing.T) {
			// Create.
			err := Assessment.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Assessment.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			r.Required = false
			err = Assessment.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Assessment.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Name != r.Name {
				t.Errorf("Different response error. Got %s, expected %s", got.Name, r.Name)
			}
			if got.Required != false {
				t.Errorf("Required should be false after update. Got %+v, expected %+v", got, r)
			}

			// Delete.
			err = Assessment.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = Assessment.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestAssessmentList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert.Must(t, Assessment.Create(&sample))
		samples[name] = sample
	}

	got, err := Assessment.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert.Must(t, Assessment.Delete(r.ID))
	}
}
