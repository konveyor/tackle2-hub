package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

// Routes
const (
	QuestionnairesRoot = "/questionnaires"
	QuestionnaireRoot  = QuestionnairesRoot + "/:" + ID
)

// QuestionnaireHandler handles Questionnaire resource routes.
type QuestionnaireHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h QuestionnaireHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("questionnaires"), Transaction)
	routeGroup.GET(QuestionnairesRoot, h.List)
	routeGroup.GET(QuestionnairesRoot+"/", h.List)
	routeGroup.POST(QuestionnairesRoot, h.Create)
	routeGroup.GET(QuestionnaireRoot, h.Get)
	routeGroup.PUT(QuestionnaireRoot, h.Update)
	routeGroup.DELETE(QuestionnaireRoot, h.Delete)
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
	var fields map[string]interface{}
	if m.Builtin() {
		fields = map[string]interface{}{
			"updateUser": updated.UpdateUser,
			"required":   updated.Required,
		}
	} else {
		fields = h.fields(updated)
	}

	db = h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result = db.Updates(fields)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

type Questionnaire struct {
	Resource     `yaml:",inline"`
	Name         string                  `json:"name" yaml:"name" binding:"required"`
	Description  string                  `json:"description" yaml:"description"`
	Required     bool                    `json:"required" yaml:"required"`
	Sections     []assessment.Section    `json:"sections" yaml:"sections" binding:"required,min=1,dive"`
	Thresholds   assessment.Thresholds   `json:"thresholds" yaml:"thresholds" binding:"required"`
	RiskMessages assessment.RiskMessages `json:"riskMessages" yaml:"riskMessages" binding:"required"`
	Builtin      bool                    `json:"builtin,omitempty" yaml:"builtin,omitempty"`
}

// With updates the resource with the model.
func (r *Questionnaire) With(m *model.Questionnaire) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Required = m.Required
	r.Builtin = m.Builtin()
	_ = json.Unmarshal(m.Sections, &r.Sections)
	_ = json.Unmarshal(m.Thresholds, &r.Thresholds)
	_ = json.Unmarshal(m.RiskMessages, &r.RiskMessages)
}

// Model builds a model.
func (r *Questionnaire) Model() (m *model.Questionnaire) {
	m = &model.Questionnaire{
		Name:        r.Name,
		Description: r.Description,
		Required:    r.Required,
	}
	m.ID = r.ID
	m.Sections, _ = json.Marshal(r.Sections)
	m.Thresholds, _ = json.Marshal(r.Thresholds)
	m.RiskMessages, _ = json.Marshal(r.RiskMessages)

	return
}
