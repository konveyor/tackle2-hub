package resource

import (
	"fmt"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Questionnaire REST resource.
type Questionnaire api.Questionnaire

// With updates the resource with the model.
func (r *Questionnaire) With(m *model.Questionnaire) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Required = m.Required
	r.Builtin = m.Builtin()
	r.Sections = []api.Section{}
	for _, s := range m.Sections {
		sect := Section{}
		sect.With(&s)
		r.Sections = append(r.Sections, api.Section(sect))
	}
	r.Thresholds = api.Thresholds(m.Thresholds)
	r.RiskMessages = api.RiskMessages(m.RiskMessages)
}

// Model builds a model.
func (r *Questionnaire) Model() (m *model.Questionnaire) {
	m = &model.Questionnaire{
		Name:        r.Name,
		Description: r.Description,
		Required:    r.Required,
	}
	m.ID = r.ID
	for _, s := range r.Sections {
		sect := Section(s)
		m.Sections = append(m.Sections, *sect.Model())
	}
	m.Thresholds = model.Thresholds(r.Thresholds)
	m.RiskMessages = model.RiskMessages(r.RiskMessages)

	return
}

// Validate performs additional validation on the questionnaire beyond binding tags.
func (r *Questionnaire) Validate() error {
	// Validate sections have unique order values
	sectionOrders := make(map[uint]bool)
	for i, section := range r.Sections {
		// Check for duplicate section order
		if sectionOrders[section.Order] {
			return &ValidationError{
				fmt.Sprintf("duplicate section order %d found", section.Order),
			}
		}
		sectionOrders[section.Order] = true

		// Validate each section has at least one question
		if len(section.Questions) == 0 {
			return &ValidationError{
				fmt.Sprintf("section %d (%s) must have at least one question", i, section.Name),
			}
		}

		// Validate questions within section
		questionOrders := make(map[uint]bool)
		for j, question := range section.Questions {
			// Check for duplicate question order within section
			if questionOrders[question.Order] {
				return &ValidationError{
					fmt.Sprintf("duplicate question order %d found in section %d (%s)", question.Order, i, section.Name),
				}
			}
			questionOrders[question.Order] = true

			// Validate question text is not empty
			if question.Text == "" {
				return &ValidationError{
					fmt.Sprintf("question %d in section %d (%s) must have text", j, i, section.Name),
				}
			}

			// Validate each question has at least one answer
			if len(question.Answers) == 0 {
				return &ValidationError{
					fmt.Sprintf("question %d (%s) in section %d (%s) must have at least one answer", j, question.Text, i, section.Name),
				}
			}

			// Validate answers within question
			answerOrders := make(map[uint]bool)
			for k, answer := range question.Answers {
				// Check for duplicate answer order within question
				if answerOrders[answer.Order] {
					return &ValidationError{
						fmt.Sprintf("duplicate answer order %d found in question %d (%s) in section %d (%s)", answer.Order, j, question.Text, i, section.Name),
					}
				}
				answerOrders[answer.Order] = true

				// Validate answer text is not empty
				if answer.Text == "" {
					return &ValidationError{
						fmt.Sprintf("answer %d in question %d (%s) in section %d (%s) must have text", k, j, question.Text, i, section.Name),
					}
				}

				// Validate risk level (already validated by binding tag, but double-check)
				validRisks := map[string]bool{"red": true, "yellow": true, "green": true, "unknown": true}
				if !validRisks[answer.Risk] {
					return &ValidationError{
						fmt.Sprintf("answer %d (%s) in question %d (%s) has invalid risk level '%s', must be one of: red, yellow, green, unknown", k, answer.Text, j, question.Text, answer.Risk),
					}
				}
			}
		}
	}

	// Validate threshold values
	if r.Thresholds.Red == 0 && r.Thresholds.Yellow == 0 && r.Thresholds.Unknown == 0 {
		return &ValidationError{
			"at least one threshold (red, yellow, or unknown) must be greater than 0",
		}
	}

	// Validate risk messages are not empty
	if r.RiskMessages.Red == "" || r.RiskMessages.Yellow == "" ||
		r.RiskMessages.Green == "" || r.RiskMessages.Unknown == "" {
		return &ValidationError{
			"all risk messages (red, yellow, green, unknown) must be provided",
		}
	}

	return nil
}
