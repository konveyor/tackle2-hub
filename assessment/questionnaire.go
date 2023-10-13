package assessment

import (
	"encoding/json"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

//
// NewQuestionnaireResolver builds a QuestionnaireResolver.
func NewQuestionnaireResolver(db *gorm.DB) (a *QuestionnaireResolver, err error) {
	a = &QuestionnaireResolver{db: db}
	a.requiredQuestionnaires = NewSet()
	err = a.cacheQuestionnaires()
	return
}

//
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

//
// Risk returns the single highest risk score for a group of assessments.
func (r *QuestionnaireResolver) Risk(assessments []model.Assessment) (risk string) {
	risk = RiskUnknown
	if len(assessments) == 0 {
		return
	}
	yellow := 0
	unknown := 0
	green := 0
	if len(assessments) > 0 {
		for _, a := range assessments {
			switch RiskLevel(&a) {
			case RiskRed:
				risk = RiskRed
				return
			case RiskYellow:
				yellow++
			case RiskGreen:
				green++
			default:
				unknown++
			}
		}
	}

	switch {
	case unknown > 0:
		risk = RiskUnknown
	case yellow > 0:
		risk = RiskYellow
	case green == len(assessments):
		risk = RiskGreen
	}

	return
}

//
// Confidence returns a total confidence score for a group of assessments.
func (r *QuestionnaireResolver) Confidence(assessments []model.Assessment) (confidence int) {
	allSections := []Section{}
	for _, a := range assessments {
		sections := []Section{}
		_ = json.Unmarshal(a.Sections, &sections)
		allSections = append(allSections, sections...)
	}
	confidence = Confidence(allSections)

	return
}

//
// Assessed returns whether a slice contains a completed assessment for each of the required
// questionnaires.
func (r *QuestionnaireResolver) Assessed(assessments []model.Assessment) (assessed bool) {
	answered := NewSet()
loop:
	for _, a := range assessments {
		if r.requiredQuestionnaires.Contains(a.QuestionnaireID) {
			sections := []Section{}
			_ = json.Unmarshal(a.Sections, &sections)
			for _, s := range sections {
				if !s.Complete() {
					continue loop
				}
			}
			answered.Add(a.QuestionnaireID)
		}
	}
	assessed = answered.Superset(r.requiredQuestionnaires, false)

	return
}
