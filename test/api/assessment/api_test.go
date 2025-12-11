package assessment

import (
	"fmt"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/api/questionnaire"
	"github.com/konveyor/tackle2-hub/test/assert"
)

func TestAssessmentCRUD(t *testing.T) {
	for _, r := range Samples {
		t.Run(fmt.Sprintf("%s for application %s", r.Questionnaire.Name, r.Application.Name), func(t *testing.T) {
			// Prepare questionnaire
			questionnaire := questionnaire.Questionnaire1
			assert.Must(t, RichClient.Questionnaire.Create(&questionnaire))
			r.Questionnaire.ID = questionnaire.ID

			// Create via parent resource.
			if r.Application.Name != "" {
				app := api.Application{Name: r.Application.Name}
				assert.Must(t, RichClient.Application.Create(&app))
				r.Application.ID = app.ID
				err := RichClient.Application.Assessment(app.ID).Create(&r)
				if err != nil {
					t.Errorf(err.Error())
				}

			}

			// Get.
			got, err := Assessment.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if assert.FlatEqual(got, r) {
				t.Errorf("Different response error. Got %v, expected %v", got, r)
			}

			// Get via parent object Application.
			gotList, err := RichClient.Application.Assessment(r.Application.ID).List()
			if err != nil {
				t.Errorf(err.Error())
			}
			found := false
			for _, gotItem := range gotList {
				if gotItem.ID == r.ID {
					found = true
				}
			}
			if !found {
				t.Errorf("Cannot find Assessment ID:%d on parent Application ID:%d", r.ID, r.Application.ID)
			}

			// Update example - select green instead of blue.
			r.Sections[0].Questions[0].Answers[2].Selected = false // blue (default)
			r.Sections[0].Questions[0].Answers[1].Selected = true  // green
			err = Assessment.Update(&r)
			if err != nil {
				t.Errorf(err.Error())
			}

			got, err = Assessment.Get(r.ID)
			if err != nil {
				t.Errorf(err.Error())
			}
			if got.Sections[0].Questions[0].Answers[2].Selected { // blue not selected
				t.Errorf("Different response error. Blue should not be selected.")
			}
			if !got.Sections[0].Questions[0].Answers[1].Selected { // green selected
				t.Errorf("Different response error. Green should be selected.")
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

			assert.Must(t, RichClient.Application.Delete(r.Application.ID))
			assert.Must(t, RichClient.Questionnaire.Delete(r.Questionnaire.ID))
		})
	}
}
