package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
)

//
// Routes
const (
	AssessmentsRoot = "/assessments"
	AssessmentRoot  = AssessmentsRoot + "/:" + ID
)

//
// AssessmentHandler handles Assessment resource routes.
type AssessmentHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h AssessmentHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("assessments"), Transaction)
	routeGroup.GET(AssessmentsRoot, h.List)
	routeGroup.GET(AssessmentsRoot+"/", h.List)
	routeGroup.GET(AssessmentRoot, h.Get)
	routeGroup.PUT(AssessmentRoot, h.Update)
	routeGroup.DELETE(AssessmentRoot, h.Delete)
}

// Get godoc
// @summary Get an assessment by ID.
// @description Get an assessment by ID.
// @tags questionnaires
// @produce json
// @success 200 {object} api.Assessment
// @router /assessments/{id} [get]
// @param id path int true "Assessment ID"
func (h AssessmentHandler) Get(ctx *gin.Context) {
	m := &model.Assessment{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := Assessment{}
	r.With(m)

	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all assessments.
// @description List all assessments.
// @tags assessments
// @produce json
// @success 200 {object} []api.Assessment
// @router /assessments [get]
func (h AssessmentHandler) List(ctx *gin.Context) {
	var list []model.Assessment
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Assessment{}
	for i := range list {
		r := Assessment{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Delete godoc
// @summary Delete an assessment.
// @description Delete an assessment.
// @tags assessments
// @success 204
// @router /assessments/{id} [delete]
// @param id path int true "Assessment ID"
func (h AssessmentHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Assessment{}
	result := h.DB(ctx).First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
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
// @summary Update an assessment.
// @description Update an assessment.
// @tags assessments
// @accept json
// @success 204
// @router /assessments/{id} [put]
// @param id path int true "Assessment ID"
// @param assessment body api.Assessment true "Assessment data"
func (h AssessmentHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Assessment{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations, "Thresholds", "RiskMessages")
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.DB(ctx).Model(m).Association("Stakeholders").Replace("Stakeholders", m.Stakeholders)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("StakeholderGroups").Replace("StakeholderGroups", m.StakeholderGroups)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

//
// Assessment REST resource.
type Assessment struct {
	Resource          `yaml:",inline"`
	Application       *Ref                 `json:"application,omitempty" yaml:",omitempty" binding:"excluded_with=Archetype"`
	Archetype         *Ref                 `json:"archetype,omitempty" yaml:",omitempty" binding:"excluded_with=Application"`
	Questionnaire     Ref                  `json:"questionnaire" binding:"required"`
	Sections          []assessment.Section `json:"sections"`
	Stakeholders      []Ref                `json:"stakeholders"`
	StakeholderGroups []Ref                `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	// read only
	Risk         string                  `json:"risk"`
	Confidence   int                     `json:"confidence"`
	Status       string                  `json:"status"`
	Thresholds   assessment.Thresholds   `json:"thresholds"`
	RiskMessages assessment.RiskMessages `json:"riskMessages" yaml:"riskMessages"`
}

//
// With updates the resource with the model.
func (r *Assessment) With(m *model.Assessment) {
	r.Resource.With(&m.Model)
	r.Questionnaire = r.ref(m.QuestionnaireID, &m.Questionnaire)
	r.Archetype = r.refPtr(m.ArchetypeID, m.Archetype)
	r.Application = r.refPtr(m.ApplicationID, m.Application)
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
	r.StakeholderGroups = []Ref{}
	for _, sg := range m.StakeholderGroups {
		ref := Ref{}
		ref.With(sg.ID, sg.Name)
		r.StakeholderGroups = append(r.StakeholderGroups, ref)
	}
	a := assessment.Assessment{}
	a.With(m)
	r.Risk = a.Risk()
	r.Confidence = a.Confidence()
	r.RiskMessages = a.RiskMessages
	r.Thresholds = a.Thresholds
	r.Sections = a.Sections
	r.Status = a.Status()
}

//
// Model builds a model.
func (r *Assessment) Model() (m *model.Assessment) {
	m = &model.Assessment{}
	m.ID = r.ID
	if r.Sections != nil {
		m.Sections, _ = json.Marshal(r.Sections)
	}
	m.QuestionnaireID = r.Questionnaire.ID
	if r.Archetype != nil {
		m.ArchetypeID = &r.Archetype.ID
	}
	if r.Application != nil {
		m.ApplicationID = &r.Application.ID
	}
	for _, ref := range r.Stakeholders {
		m.Stakeholders = append(
			m.Stakeholders,
			model.Stakeholder{
				Model: model.Model{ID: ref.ID},
			})
	}
	for _, ref := range r.StakeholderGroups {
		m.StakeholderGroups = append(
			m.StakeholderGroups,
			model.StakeholderGroup{
				Model: model.Model{ID: ref.ID},
			})
	}
	return
}
