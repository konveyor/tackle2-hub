package assessment

import (
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

// NewQuestionnaireResolver builds a QuestionnaireResolver.
func NewQuestionnaireResolver(db *gorm.DB) (a *QuestionnaireResolver, err error) {
	a = &QuestionnaireResolver{db: db}
	a.requiredQuestionnaires = NewSet()
	err = a.cacheQuestionnaires()
	return
}

// QuestionnaireResolver resolves questionnaire logic.
type QuestionnaireResolver struct {
	db                     *gorm.DB
	requiredQuestionnaires Set
}

func (r *QuestionnaireResolver) cacheQuestionnaires() (err error) {
	if r.requiredQuestionnaires.Size() > 0 {
		return
	}

	questionnaires := []model.Questionnaire{}
	result := r.db.Find(&questionnaires, "required = ?", true)
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
