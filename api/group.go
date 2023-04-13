package api

import (
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
)

//
// Routes
const (
	StakeholderGroupsRoot = "/stakeholdergroups"
	StakeholderGroupRoot  = StakeholderGroupsRoot + "/:" + ID
)

//
// StakeholderGroupHandler handles stakeholder group routes.
type StakeholderGroupHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h StakeholderGroupHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("stakeholdergroups"), Transaction)
	routeGroup.GET(StakeholderGroupsRoot, h.List)
	routeGroup.GET(StakeholderGroupsRoot+"/", h.List)
	routeGroup.POST(StakeholderGroupsRoot, h.Create)
	routeGroup.GET(StakeholderGroupRoot, h.Get)
	routeGroup.PUT(StakeholderGroupRoot, h.Update)
	routeGroup.DELETE(StakeholderGroupRoot, h.Delete)
}

// Get godoc
// @summary Get a stakeholder group by ID.
// @description Get a stakeholder group by ID.
// @tags stakeholdergroups
// @produce json
// @success 200 {object} api.StakeholderGroup
// @router /stakeholdergroups/{id} [get]
// @param id path string true "Stakeholder Group ID"
func (h StakeholderGroupHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.StakeholderGroup{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r := StakeholderGroup{}
	r.With(m)

	h.Render(ctx, http.StatusOK, m)
}

// List godoc
// @summary List all stakeholder groups.
// @description List all stakeholder groups.
// @tags stakeholdergroups
// @produce json
// @success 200 {object} []api.StakeholderGroup
// @router /stakeholdergroups [get]
func (h StakeholderGroupHandler) List(ctx *gin.Context) {
	var list []model.StakeholderGroup
	db := h.preLoad(h.Paginated(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []StakeholderGroup{}
	for i := range list {
		r := StakeholderGroup{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Render(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a stakeholder group.
// @description Create a stakeholder group.
// @tags stakeholdergroups
// @accept json
// @produce json
// @success 201 {object} api.StakeholderGroup
// @router /stakeholdergroups [post]
// @param stakeholder_group body api.StakeholderGroup true "Stakeholder Group data"
func (h StakeholderGroupHandler) Create(ctx *gin.Context) {
	r := &StakeholderGroup{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.BaseHandler.CurrentUser(ctx)
	result := h.DB(ctx).Create(m)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	r.With(m)

	h.Render(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a stakeholder group.
// @description Delete a stakeholder group.
// @tags stakeholdergroups
// @success 204
// @router /stakeholdergroups/{id} [delete]
// @param id path string true "Stakeholder Group ID"
func (h StakeholderGroupHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.StakeholderGroup{}
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

// Update godoc
// @summary Update a stakeholder group.
// @description Update a stakeholder group.
// @tags stakeholdergroups
// @accept json
// @success 204
// @router /stakeholdergroups/{id} [put]
// @param id path string true "Stakeholder Group ID"
// @param stakeholder_group body api.StakeholderGroup true "Stakeholder Group data"
func (h StakeholderGroupHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &StakeholderGroup{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.BaseHandler.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	result := db.Updates(h.fields(m))
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Stakeholders").Replace(m.Stakeholders)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("MigrationWaves").Replace(m.MigrationWaves)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

//
// StakeholderGroup REST resource.
type StakeholderGroup struct {
	Resource
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	Stakeholders   []Ref  `json:"stakeholders"`
	MigrationWaves []Ref  `json:"migrationWaves"`
}

//
// With updates the resource with the model.
func (r *StakeholderGroup) With(m *model.StakeholderGroup) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Stakeholders = []Ref{}
	for _, s := range m.Stakeholders {
		ref := Ref{}
		ref.With(s.ID, s.Name)
		r.Stakeholders = append(r.Stakeholders, ref)
	}
	r.MigrationWaves = []Ref{}
	for _, w := range m.MigrationWaves {
		ref := Ref{}
		ref.With(w.ID, w.Name)
		r.MigrationWaves = append(r.MigrationWaves, ref)
	}
}

//
// Model builds a model.
func (r *StakeholderGroup) Model() (m *model.StakeholderGroup) {
	m = &model.StakeholderGroup{
		Name:        r.Name,
		Description: r.Description,
	}
	m.ID = r.ID
	for _, s := range r.Stakeholders {
		m.Stakeholders = append(m.Stakeholders, model.Stakeholder{Model: model.Model{ID: s.ID}})
	}
	for _, w := range r.MigrationWaves {
		m.MigrationWaves = append(m.MigrationWaves, model.MigrationWave{Model: model.Model{ID: w.ID}})
	}
	return
}
