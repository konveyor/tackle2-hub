package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/assessment"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	ArchetypesRoot           = "/archetypes"
	ArchetypeRoot            = ArchetypesRoot + "/:" + ID
	ArchetypeAssessmentsRoot = ArchetypeRoot + "/assessments"
)

//
// ArchetypeHandler handles Archetype resource routes.
type ArchetypeHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h ArchetypeHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("archetypes"), Transaction)
	routeGroup.GET(ArchetypesRoot, h.List)
	routeGroup.POST(ArchetypesRoot, h.Create)
	routeGroup.GET(ArchetypeRoot, h.Get)
	routeGroup.PUT(ArchetypeRoot, h.Update)
	routeGroup.DELETE(ArchetypeRoot, h.Delete)
	// Assessments
	routeGroup = e.Group("/")
	routeGroup.Use(Required("archetypes.assessments"))
	routeGroup.GET(ArchetypeAssessmentsRoot, h.AssessmentList)
	routeGroup.POST(ArchetypeAssessmentsRoot, h.AssessmentCreate)
}

// Get godoc
// @summary Get an archetype by ID.
// @description Get an archetype by ID.
// @tags archetypes
// @produce json
// @success 200 {object} api.Archetypes
// @router /archetypes/{id} [get]
// @param id path string true "Archetype ID"
func (h ArchetypeHandler) Get(ctx *gin.Context) {
	m := &model.Archetype{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	membership := assessment.NewMembershipResolver(h.DB(ctx))
	applications, err := membership.Applications(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := Archetype{}
	r.With(m)
	r.WithApplications(applications)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all archetypes.
// @description List all archetypes.
// @tags archetypes
// @produce json
// @success 200 {object} []api.Archetype
// @router /archetypes [get]
func (h ArchetypeHandler) List(ctx *gin.Context) {
	var list []model.Archetype
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	membership := assessment.NewMembershipResolver(h.DB(ctx))
	resources := []Archetype{}
	for i := range list {
		r := Archetype{}
		applications, err := membership.Applications(&list[i])
		if err != nil {
			_ = ctx.Error(err)
			return
		}
		r.With(&list[i])
		r.WithApplications(applications)
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create an archetype.
// @description Create an archetype.
// @tags archetypes
// @accept json
// @produce json
// @success 200 {object} api.Archetype
// @router /archetypes [post]
// @param archetype body api.Archetype true "Archetype data"
func (h ArchetypeHandler) Create(ctx *gin.Context) {
	r := &Archetype{}
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

	archetypes := []model.Archetype{}
	db := h.preLoad(h.DB(ctx), "Tags", "CriteriaTags")
	result = db.Find(&archetypes)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
	}

	membership := assessment.NewMembershipResolver(h.DB(ctx))
	applications, err := membership.Applications(m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r.With(m)
	r.WithApplications(applications)
	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an archetype.
// @description Delete an archetype.
// @tags archetypes
// @success 204
// @router /archetypes/{id} [delete]
// @param id path string true "Archetype ID"
func (h ArchetypeHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Archetype{}
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
// @summary Update an archetype.
// @description Update an archetype.
// @tags archetypes
// @accept json
// @success 204
// @router /archetypes/{id} [put]
// @param id path string true "Archetype ID"
// @param archetype body api.Archetype true "Archetype data"
func (h ArchetypeHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Archetype{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
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
	err = h.DB(ctx).Model(m).Association("CriteriaTags").Replace("CriteriaTags", m.CriteriaTags)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Model(m).Association("Tags").Replace("Tags", m.Tags)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// AssessmentList godoc
// @summary List the assessments of an archetype.
// @description List the assessments of an archetype.
// @tags archetypes
// @success 200 {object} []api.Assessment
// @router /archetypes/{id}/assessments [get]
// @param id path int true "Archetype ID"
func (h ArchetypeHandler) AssessmentList(ctx *gin.Context) {
	m := &model.Archetype{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	resources := []Assessment{}
	for i := range m.Assessments {
		r := Assessment{}
		r.With(&m.Assessments[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// AssessmentCreate godoc
// @summary Create an archetype assessment.
// @description Create an archetype assessment.
// @tags archetypes
// @accept json
// @produce json
// @success 201 {object} api.Assessment
// @router /archetypes/{id}/assessments [post]
// @param assessment body api.Assessment true "Assessment data"
func (h ArchetypeHandler) AssessmentCreate(ctx *gin.Context) {
	archetype := &model.Archetype{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(archetype, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := &Assessment{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.Archetype = &Ref{ID: id}
	r.Application = nil
	q := &model.Questionnaire{}
	result = h.DB(ctx).First(q, r.Questionnaire.ID)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	m := r.Model()
	m.Sections = q.Sections
	m.Thresholds = q.Thresholds
	m.RiskMessages = q.RiskMessages
	m.CreateUser = h.CurrentUser(ctx)

	resolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	assessment.PrepareForArchetype(resolver, archetype, m)
	result = h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

//
// Archetype REST resource.
type Archetype struct {
	Resource
	Name              string `json:"name" yaml:"name"`
	Description       string `json:"description" yaml:"description"`
	Comments          string `json:"comments" yaml:"comments"`
	Tags              []Ref  `json:"tags" yaml:"tags"`
	CriteriaTags      []Ref  `json:"criteriaTags" yaml:"criteriaTags"`
	Stakeholders      []Ref  `json:"stakeholders" yaml:"stakeholders"`
	StakeholderGroups []Ref  `json:"stakeholderGroups" yaml:"stakeholderGroups"`
	Applications      []Ref  `json:"applications" yaml:"applications"`
	Assessments       []Ref  `json:"assessments" yaml:"assessments"`
	Review            *Ref   `json:"review"`
}

//
// With updates the resource with the model.
func (r *Archetype) With(m *model.Archetype) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Comments = m.Comments
	r.Tags = []Ref{}
	for _, t := range m.Tags {
		r.Tags = append(r.Tags, r.ref(t.ID, &t))
	}
	r.CriteriaTags = []Ref{}
	for _, t := range m.CriteriaTags {
		r.CriteriaTags = append(r.CriteriaTags, r.ref(t.ID, &t))
	}
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		r.Stakeholders = append(r.Stakeholders, r.ref(s.ID, &s))
	}
	r.StakeholderGroups = []Ref{}
	for _, g := range m.StakeholderGroups {
		r.StakeholderGroups = append(r.StakeholderGroups, r.ref(g.ID, &g))
	}
	r.Assessments = []Ref{}
	for _, a := range m.Assessments {
		r.Assessments = append(r.Assessments, r.ref(a.ID, &a))
	}
	if m.Review != nil {
		ref := &Ref{}
		ref.With(m.Review.ID, "")
		r.Review = ref
	}
}

//
// WithApplications updates the Archetype resource with the applications.
func (r *Archetype) WithApplications(apps []model.Application) {
	for i := range apps {
		ref := Ref{}
		ref.With(apps[i].ID, apps[i].Name)
		r.Applications = append(r.Applications, ref)
	}
}

//
// Model builds a model from the resource.
func (r *Archetype) Model() (m *model.Archetype) {
	m = &model.Archetype{
		Name:        r.Name,
		Description: r.Description,
		Comments:    r.Comments,
	}
	m.ID = r.ID
	for _, ref := range r.Tags {
		m.Tags = append(
			m.Tags,
			model.Tag{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.CriteriaTags {
		m.CriteriaTags = append(
			m.CriteriaTags,
			model.Tag{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.Stakeholders {
		m.Stakeholders = append(
			m.Stakeholders,
			model.Stakeholder{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}
	for _, ref := range r.StakeholderGroups {
		m.StakeholderGroups = append(
			m.StakeholderGroups,
			model.StakeholderGroup{
				Model: model.Model{
					ID: ref.ID,
				},
			})
	}

	return
}
