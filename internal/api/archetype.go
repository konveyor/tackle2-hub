package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/assessment"
	"github.com/konveyor/tackle2-hub/internal/metrics"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ArchetypeHandler handles Archetype resource routes.
type ArchetypeHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h ArchetypeHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("archetypes"), Transaction)
	routeGroup.GET(api.ArchetypesRoute, h.List)
	routeGroup.POST(api.ArchetypesRoute, h.Create)
	routeGroup.GET(api.ArchetypeRoute, h.Get)
	routeGroup.PUT(api.ArchetypeRoute, h.Update)
	routeGroup.DELETE(api.ArchetypeRoute, h.Delete)
	// Assessments
	routeGroup = e.Group("/")
	routeGroup.Use(Required("archetypes.assessments"))
	routeGroup.GET(api.ArchetypeAssessmentsRoute, h.AssessmentList)
	routeGroup.POST(api.ArchetypeAssessmentsRoute, h.AssessmentCreate)
}

// Get godoc
// @summary Get an archetype by ID.
// @description Get an archetype by ID.
// @tags archetypes
// @produce json
// @success 200 {object} api.Archetype
// @router /archetypes/{id} [get]
// @param id path int true "Archetype ID"
func (h ArchetypeHandler) Get(ctx *gin.Context) {
	m := &model.Archetype{}
	id := h.pk(ctx)
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	db = db.Preload("Profiles.AnalysisProfile")
	db = db.Preload("Profiles.Generators.Generator")
	db = db.Preload("Profiles.Generators", func(db *gorm.DB) *gorm.DB {
		return db.Order("`Index`")
	})
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	memberResolver, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	questResolver, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	tagResolver, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
	}
	resolver := assessment.NewArchetypeResolver(m, tagResolver, memberResolver, questResolver)
	r := Archetype{}
	r.With(m)
	err = r.WithResolver(resolver)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
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
	db := h.DB(ctx)
	db = db.Preload(clause.Associations)
	db = db.Preload("Profiles.AnalysisProfile")
	db = db.Preload("Profiles.Generators.Generator")
	db = db.Preload("Profiles.Generators", func(db *gorm.DB) *gorm.DB {
		return db.Order("`Index`")
	})
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	questionnaires, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	tags, err := assessment.NewTagResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
	}
	membership, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []Archetype{}
	for i := range list {
		m := &list[i]
		resolver := assessment.NewArchetypeResolver(m, tags, membership, questionnaires)
		r := Archetype{}
		r.With(m)
		err = r.WithResolver(resolver)
		if err != nil {
			_ = ctx.Error(err)
			return
		}
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
	result := h.DB(ctx).Omit(clause.Associations).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.adjustProfileIds(ctx, m)
	if err != nil {
		_ = ctx.Error(err)
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
	err = h.Association(ctx, "Profiles").Owner(true).Replace(m, m.Profiles)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.updateGenerators(ctx, m)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	archetypes := []model.Archetype{}
	db := h.preLoad(h.DB(ctx), "Tags", "CriteriaTags")
	result = db.Find(&archetypes)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
	}

	membership, err := assessment.NewMembershipResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	questionnaires, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resolver := assessment.NewArchetypeResolver(m, nil, membership, questionnaires)
	r.With(m)
	err = r.WithResolver(resolver)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete an archetype.
// @description Delete an archetype.
// @tags archetypes
// @success 204
// @router /archetypes/{id} [delete]
// @param id path int true "Archetype ID"
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
// @param id path int true "Archetype ID"
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
	result := db.Save(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	err = h.adjustProfileIds(ctx, m)
	if err != nil {
		_ = ctx.Error(err)
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
	err = h.Association(ctx, "Profiles").Owner(true).Replace(m, m.Profiles)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.updateGenerators(ctx, m)
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
	db := h.preLoad(
		h.DB(ctx),
		clause.Associations,
		"Assessments.Stakeholders",
		"Assessments.StakeholderGroups",
		"Assessments.Questionnaire")
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
// @param id path int true "Archetype ID"
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
	r.Archetype = &resource.Ref{ID: id}
	r.Application = nil
	q := &model.Questionnaire{}
	result = h.DB(ctx).First(q, r.Questionnaire.ID)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	m := r.Model()
	m.Thresholds = q.Thresholds
	m.RiskMessages = q.RiskMessages
	m.CreateUser = h.CurrentUser(ctx)
	// if sections aren't empty that indicates that this assessment is being
	// created "as-is" and should not have its sections populated or autofilled.
	newAssessment := false
	if len(m.Sections) == 0 {
		m.Sections = q.Sections
		resolver, rErr := assessment.NewTagResolver(h.DB(ctx))
		if rErr != nil {
			_ = ctx.Error(rErr)
			return
		}
		assessment.PrepareForArchetype(resolver, archetype, m)
		newAssessment = true
	}
	result = h.DB(ctx).Omit(clause.Associations).Create(m)
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
	if newAssessment {
		metrics.AssessmentsInitiated.Inc()
	}

	r.With(m)
	h.Respond(ctx, http.StatusCreated, r)
}

// adjustProfileIds adjust profile ids as needed.
// Protect against:
// - creating profiles with explicit ids.
// - transferring a profile owned by another.
func (h ArchetypeHandler) adjustProfileIds(ctx *gin.Context, m *model.Archetype) (err error) {
	var owned []model.TargetProfile
	db := h.DB(ctx)
	db = db.Where("ArchetypeID", m.ID)
	err = db.Find(&owned).Error
	if err != nil {
		return
	}
	ids := make(map[uint]uint)
	for _, p := range owned {
		ids[p.ID] = m.ID
	}
	for i := range m.Profiles {
		p := &m.Profiles[i]
		if _, found := ids[p.ID]; !found {
			p.ID = 0
		}
	}
	return
}

// updateGenerators replaces generators.
func (h ArchetypeHandler) updateGenerators(ctx *gin.Context, m *model.Archetype) (err error) {
	for i := range m.Profiles {
		p := &m.Profiles[i]
		db := h.DB(ctx)
		db = db.Where("TargetProfileId", p.ID)
		err = db.Delete(&model.ProfileGenerator{}).Error
		if err != nil {
			return
		}
		for index := range p.Generators {
			db := h.DB(ctx)
			g := &p.Generators[index]
			g.TargetProfileID = p.ID
			g.TargetProfile.ID = p.ID
			g.Generator.ID = g.GeneratorID
			g.Index = index
			err = db.Create(g).Error
			if err != nil {
				return
			}
		}
	}
	return
}

// TargetProfile REST resource.
type TargetProfile = resource.TargetProfile

// Archetype REST resource.
type Archetype = resource.Archetype
