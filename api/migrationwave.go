package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

//
// Routes
const (
	MigrationWavesRoot = "/migrationwaves"
	MigrationWaveRoot  = MigrationWavesRoot + "/:" + ID
)

//
// MigrationWaveHandler handles Migration Wave resource routes.
type MigrationWaveHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h MigrationWaveHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("migrationwaves"), Transaction)
	routeGroup.GET(MigrationWavesRoot, h.List)
	routeGroup.GET(MigrationWavesRoot+"/", h.List)
	routeGroup.GET(MigrationWaveRoot, h.Get)
	routeGroup.POST(MigrationWavesRoot, h.Create)
	routeGroup.DELETE(MigrationWaveRoot, h.Delete)
	routeGroup.PUT(MigrationWaveRoot, h.Update)
}

// Get godoc
// @summary Get aa migration wave by ID.
// @description Get a migration wave by ID.
// @tags migrationwaves
// @produce json
// @success 200 {object} api.MigrationWave
// @router /migrationwaves/{id} [get]
// @param id path int true "Migration Wave ID"
func (h MigrationWaveHandler) Get(ctx *gin.Context) {
	m := &model.MigrationWave{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := MigrationWave{}
	r.With(m)

	h.Render(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all migration waves.
// @description List all migration waves.
// @tags migrationwaves
// @produce json
// @success 200 {object} []api.MigrationWave
// @router /migrationwaves [get]
func (h MigrationWaveHandler) List(ctx *gin.Context) {
	var list []model.MigrationWave
	db := h.preLoad(h.Paginated(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []MigrationWave{}
	for i := range list {
		r := MigrationWave{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Render(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a migration wave.
// @description Create a migration wave.
// @tags migrationwaves
// @accept json
// @produce json
// @success 201 {object} api.MigrationWave
// @router /migrationwaves [post]
// @param migrationwave body api.MigrationWave true "Migration Wave data"
func (h MigrationWaveHandler) Create(ctx *gin.Context) {
	r := &MigrationWave{}
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

	h.Render(ctx, http.StatusCreated, r)
}

// Update godoc
// @summary Update a migration wave.
// @description Update a migration wave.
// @tags migrationwaves
// @accept json
// @success 204
// @router /migrationwaves/{id} [put]
// @param id path int true "MigrationWave id"
// @param migrationWave body api.MigrationWave true "MigrationWave data"
func (h MigrationWaveHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &MigrationWave{}
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
	err = h.DB(ctx).Model(m).Association("Applications").Replace("Applications", m.Applications)
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

	ctx.Status(http.StatusNoContent)
}

// Delete godoc
// @summary Delete a migration wave.
// @description Delete a migration wave.
// @tags migrationwaves
// @success 204
// @router /migrationwaves/{id} [delete]
// @param id path int true "MigrationWave id"
func (h MigrationWaveHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.MigrationWave{}
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

	ctx.Status(http.StatusNoContent)
}

//
// MigrationWave REST Resource
type MigrationWave struct {
	Resource
	Name              string    `json:"name"`
	StartDate         time.Time `json:"startDate"`
	EndDate           time.Time `json:"endDate"`
	Applications      []Ref     `json:"applications"`
	Stakeholders      []Ref     `json:"stakeholders"`
	StakeholderGroups []Ref     `json:"stakeholderGroups"`
}

//
// With updates the resource using the model.
func (r *MigrationWave) With(m *model.MigrationWave) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.StartDate = m.StartDate
	r.EndDate = m.EndDate
	r.Applications = []Ref{}
	for _, app := range m.Applications {
		ref := Ref{}
		ref.With(app.ID, app.Name)
		r.Applications = append(r.Applications, ref)
	}
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
}

//
// Model builds a model.
func (r *MigrationWave) Model() (m *model.MigrationWave) {
	m = &model.MigrationWave{
		Name:      r.Name,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
	}
	m.ID = r.ID
	for _, ref := range r.Applications {
		m.Applications = append(
			m.Applications,
			model.Application{
				Model: model.Model{ID: ref.ID},
			})
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
