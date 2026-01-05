package assessment

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"gorm.io/gorm"
)

// NewQuestionnaireResolver builds a QuestionnaireResolver.
func NewQuestionnaireResolver(db *gorm.DB) (a *QuestionnaireResolver, err error) {
	a = &QuestionnaireResolver{}
	a.requiredQuestionnaires = NewSet()
	err = a.cacheQuestionnaires(db)
	return
}

// QuestionnaireResolver resolves questionnaire logic.
type QuestionnaireResolver struct {
	requiredQuestionnaires Set
}

func (r *QuestionnaireResolver) cacheQuestionnaires(db *gorm.DB) (err error) {
	if r.requiredQuestionnaires.Size() > 0 {
		return
	}

	questionnaires := []model.Questionnaire{}
	result := db.Find(&questionnaires, "required = ?", true)
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}

	for _, q := range questionnaires {
		r.requiredQuestionnaires.Add(q.ID)
	}

	return
}

// Required returns whether a questionnaire is required.
func (r *QuestionnaireResolver) Required(id uint) (required bool) {
	return r.requiredQuestionnaires.Contains(id)
}

// Assessed returns whether a slice contains a completed assessment for each of the required
// questionnaires.
func (r *QuestionnaireResolver) Assessed(assessments []Assessment) (assessed bool) {
	if r.requiredQuestionnaires.Size() == 0 {
		return false
	}
	answered := NewSet()
	for _, a := range assessments {
		if r.requiredQuestionnaires.Contains(a.QuestionnaireID) {
			if a.Complete() {
				answered.Add(a.QuestionnaireID)
			}
		}
	}
	assessed = answered.Superset(r.requiredQuestionnaires, false)

	return
}
