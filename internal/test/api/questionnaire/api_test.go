package questionnaire

import (
	"testing"

	assert2 "github.com/konveyor/tackle2-hub/internal/test/assert"
)

func TestQuestionnaireCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(r.Name, func(t *testing.T) {
			// Create.
			err := Questionnaire.Create(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			// Get.
			got, err := Questionnaire.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert2.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Update.
			r.Name = "Updated " + r.Name
			r.Required = false
			err = Questionnaire.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Questionnaire.Get(r.ID)
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
			err = Questionnaire.Delete(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}

			_, err = Questionnaire.Get(r.ID)
			if err == nil {
				t.Errorf("Resource exits, but should be deleted: %v", r)
			}
		})
	}
}

func TestQuestionnaireList(t *testing.T) {
	samples := Samples

	for name := range samples {
		sample := samples[name]
		assert2.Must(t, Questionnaire.Create(&sample))
		samples[name] = sample
	}

	got, err := Questionnaire.List()
	if err != nil {
		t.Errorf(err.Error())
	}
	if assert2.FlatEqual(got, &samples) {
		t.Errorf("Different response error. Got %v, expected %v", got, samples)
	}

	for _, r := range samples {
		assert2.Must(t, Questionnaire.Delete(r.ID))
	}
}
