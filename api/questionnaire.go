package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

// Routes
const (
	QuestionnairesRoute = "/questionnaires"
	QuestionnaireRoute  = QuestionnairesRoute + "/:" + ID
)

// QuestionnaireHandler handles Questionnaire resource routes.
type QuestionnaireHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h QuestionnaireHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("questionnaires"), Transaction)
	routeGroup.GET(QuestionnairesRoute, h.List)
	routeGroup.GET(QuestionnairesRoute+"/", h.List)
	routeGroup.POST(QuestionnairesRoute, h.Create)
	routeGroup.GET(QuestionnaireRoute, h.Get)
	routeGroup.PUT(QuestionnaireRoute, h.Update)
	routeGroup.DELETE(QuestionnaireRoute, h.Delete)
}

// Get godoc
// @summary Get a questionnaire by ID.
// @description Get a questionnaire by ID.
// @tags questionnaires
// @produce json
// @success 200 {object} api.Questionnaire
// @router /questionnaires/{id} [get]
// @param id path int true "Questionnaire ID"
func (h QuestionnaireHandler) Get(ctx *gin.Context) {
	m := &model.Questionnaire{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Questionnaire{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all questionnaires.
// @description List all questionnaires.
// @tags questionnaires
// @produce json
// @success 200 {object} []api.Questionnaire
// @router /questionnaires [get]
func (h QuestionnaireHandler) List(ctx *gin.Context) {
	var list []model.Questionnaire
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Questionnaire{}
	for i := range list {
		r := Questionnaire{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a questionnaire.
// @description Create a questionnaire.
// @tags questionnaires
// @accept json
// @produce json
// @success 200 {object} api.Questionnaire
// @router /questionnaires [post]
// @param questionnaire body api.Questionnaire true "Questionnaire data"
func (h QuestionnaireHandler) Create(ctx *gin.Context) {
	r := &Questionnaire{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	// Additional questionaire fields validation
	err = r.Validate()
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a questionnaire.
// @description Delete a questionnaire.
// @tags questionnaires
// @success 204
// @router /questionnaires/{id} [delete]
// @param id path int true "Questionnaire ID"
func (h QuestionnaireHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Questionnaire{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	if m.Builtin() {
		h.Status(ctx, http.StatusForbidden)
		return
	}
	result = h.DB(ctx).Delete(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Update godoc
// @summary Update a questionnaire.
// @description Update a questionnaire. If the Questionnaire
// @description is builtin, only its "required" field can be changed
// @description and all other fields will be ignored.
// @tags questionnaires
// @accept json
// @success 204
// @router /questionnaires/{id} [put]
// @param id path int true "Questionnaire ID"
// @param questionnaire body api.Questionnaire true "Questionnaire data"
func (h QuestionnaireHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Questionnaire{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := &model.Questionnaire{}
	db := h.DB(ctx)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	updated := r.Model()
	updated.ID = id
	updated.UpdateUser = h.CurrentUser(ctx)
	if m.Builtin() {
		m.UpdateUser = updated.UpdateUser
		m.Required = updated.Required
	} else {
		// Additional validation for non-builtin questionnaires fields
		err = r.Validate()
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		m = updated
	}

	db = h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result = db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

type Questionnaire struct {
	Resource     `yaml:",inline"`
	Name         string       `json:"name" yaml:"name" binding:"required"`
	Description  string       `json:"description" yaml:"description"`
	Required     bool         `json:"required" yaml:"required"`
	Sections     []Section    `json:"sections" yaml:"sections" binding:"required,min=1,dive"`
	Thresholds   Thresholds   `json:"thresholds" yaml:"thresholds" binding:"required"`
	RiskMessages RiskMessages `json:"riskMessages" yaml:"riskMessages" binding:"required"`
	Builtin      bool         `json:"builtin,omitempty" yaml:"builtin,omitempty"`
}

// With updates the resource with the model.
func (r *Questionnaire) With(m *model.Questionnaire) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Required = m.Required
	r.Builtin = m.Builtin()
	r.Sections = []Section{}
	for _, s := range m.Sections {
		r.Sections = append(r.Sections, Section(s))
	}
	r.Thresholds = Thresholds(m.Thresholds)
	r.RiskMessages = RiskMessages(m.RiskMessages)
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
		m.Sections = append(m.Sections, model.Section(s))
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
			return &BadRequestError{
				fmt.Sprintf("duplicate section order %d found", section.Order),
			}
		}
		sectionOrders[section.Order] = true

		// Validate each section has at least one question
		if len(section.Questions) == 0 {
			return &BadRequestError{
				fmt.Sprintf("section %d (%s) must have at least one question", i, section.Name),
			}
		}

		// Validate questions within section
		questionOrders := make(map[uint]bool)
		for j, question := range section.Questions {
			// Check for duplicate question order within section
			if questionOrders[question.Order] {
				return &BadRequestError{
					fmt.Sprintf("duplicate question order %d found in section %d (%s)", question.Order, i, section.Name),
				}
			}
			questionOrders[question.Order] = true

			// Validate question text is not empty
			if question.Text == "" {
				return &BadRequestError{
					fmt.Sprintf("question %d in section %d (%s) must have text", j, i, section.Name),
				}
			}

			// Validate each question has at least one answer
			if len(question.Answers) == 0 {
				return &BadRequestError{
					fmt.Sprintf("question %d (%s) in section %d (%s) must have at least one answer", j, question.Text, i, section.Name),
				}
			}

			// Validate answers within question
			answerOrders := make(map[uint]bool)
			for k, answer := range question.Answers {
				// Check for duplicate answer order within question
				if answerOrders[answer.Order] {
					return &BadRequestError{
						fmt.Sprintf("duplicate answer order %d found in question %d (%s) in section %d (%s)", answer.Order, j, question.Text, i, section.Name),
					}
				}
				answerOrders[answer.Order] = true

				// Validate answer text is not empty
				if answer.Text == "" {
					return &BadRequestError{
						fmt.Sprintf("answer %d in question %d (%s) in section %d (%s) must have text", k, j, question.Text, i, section.Name),
					}
				}

				// Validate risk level (already validated by binding tag, but double-check)
				validRisks := map[string]bool{"red": true, "yellow": true, "green": true, "unknown": true}
				if !validRisks[answer.Risk] {
					return &BadRequestError{
						fmt.Sprintf("answer %d (%s) in question %d (%s) has invalid risk level '%s', must be one of: red, yellow, green, unknown", k, answer.Text, j, question.Text, answer.Risk),
					}
				}
			}
		}
	}

	// Validate threshold values
	if r.Thresholds.Red == 0 && r.Thresholds.Yellow == 0 && r.Thresholds.Unknown == 0 {
		return &BadRequestError{
			"at least one threshold (red, yellow, or unknown) must be greater than 0",
		}
	}

	// Validate risk messages are not empty
	if r.RiskMessages.Red == "" || r.RiskMessages.Yellow == "" ||
		r.RiskMessages.Green == "" || r.RiskMessages.Unknown == "" {
		return &BadRequestError{
			"all risk messages (red, yellow, green, unknown) must be provided",
		}
	}

	return nil
}
