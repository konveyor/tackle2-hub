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
	StakeholdersRoot = "/stakeholders"
	StakeholderRoot  = StakeholdersRoot + "/:" + ID
)

//
// StakeholderHandler handles stakeholder routes.
type StakeholderHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h StakeholderHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("stakeholders"), Transaction)
	routeGroup.GET(StakeholdersRoot, h.List)
	routeGroup.GET(StakeholdersRoot+"/", h.List)
	routeGroup.POST(StakeholdersRoot, h.Create)
	routeGroup.GET(StakeholderRoot, h.Get)
	routeGroup.PUT(StakeholderRoot, h.Update)
	routeGroup.DELETE(StakeholderRoot, h.Delete)
}

// Get godoc
// @summary Get a stakeholder by ID.
// @description Get a stakeholder by ID.
// @tags stakeholders
// @produce json
// @success 200 {object} api.Stakeholder
// @router /stakeholders/{id} [get]
// @param id path string true "Stakeholder ID"
func (h StakeholderHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Stakeholder{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	resource := Stakeholder{}
	resource.With(m)
	h.Render(ctx, http.StatusOK, resource)
}

// List godoc
// @summary List all stakeholders.
// @description List all stakeholders.
// @tags stakeholders
// @produce json
// @success 200 {object} []api.Stakeholder
// @router /stakeholders [get]
func (h StakeholderHandler) List(ctx *gin.Context) {
	var list []model.Stakeholder
	db := h.preLoad(h.Paginated(ctx), clause.Associations)
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Stakeholder{}
	for i := range list {
		r := Stakeholder{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Render(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a stakeholder.
// @description Create a stakeholder.
// @tags stakeholders
// @accept json
// @produce json
// @success 201 {object} api.Stakeholder
// @router /stakeholders [post]
// @param stakeholder body api.Stakeholder true "Stakeholder data"
func (h StakeholderHandler) Create(ctx *gin.Context) {
	r := &Stakeholder{}
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
// @summary Delete a stakeholder.
// @description Delete a stakeholder.
// @tags stakeholders
// @success 204
// @router /stakeholders/{id} [delete]
// @param id path string true "Stakeholder ID"
func (h StakeholderHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Stakeholder{}
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
// @summary Update a stakeholder.
// @description Update a stakeholder.
// @tags stakeholders
// @accept json
// @success 204
// @router /stakeholders/{id} [put]
// @param id path string true "Stakeholder ID"
// @param stakeholder body api.Stakeholder true "Stakeholder data"
func (h StakeholderHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Stakeholder{}
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
	err = db.Association("Groups").Replace(m.Groups)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Owns").Replace(m.Owns)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	db = h.DB(ctx).Model(m)
	err = db.Association("Contributes").Replace(m.Contributes)
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
// Stakeholder REST resource.
type Stakeholder struct {
	Resource
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email" binding:"required"`
	Groups           []Ref  `json:"stakeholderGroups"`
	BusinessServices []Ref  `json:"businessServices"`
	JobFunction      *Ref   `json:"jobFunction"`
	Owns             []Ref  `json:"owns"`
	Contributes      []Ref  `json:"contributes"`
	MigrationWaves   []Ref  `json:"migrationWaves"`
}

//
// With updates the resource with the model.
func (r *Stakeholder) With(m *model.Stakeholder) {
	r.Resource.With(&m.Model)
	r.Name = m.Name
	r.Email = m.Email
	r.JobFunction = r.refPtr(m.JobFunctionID, m.JobFunction)
	r.Groups = []Ref{}
	for _, g := range m.Groups {
		ref := Ref{}
		ref.With(g.ID, g.Name)
		r.Groups = append(r.Groups, ref)
	}
	r.BusinessServices = []Ref{}
	for _, b := range m.BusinessServices {
		ref := Ref{}
		ref.With(b.ID, b.Name)
		r.BusinessServices = append(r.BusinessServices, ref)
	}
	r.Owns = []Ref{}
	for _, o := range m.Owns {
		ref := Ref{}
		ref.With(o.ID, o.Name)
		r.Owns = append(r.Owns, ref)
	}
	r.Contributes = []Ref{}
	for _, c := range m.Contributes {
		ref := Ref{}
		ref.With(c.ID, c.Name)
		r.Contributes = append(r.Contributes, ref)
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
func (r *Stakeholder) Model() (m *model.Stakeholder) {
	m = &model.Stakeholder{
		Name:  r.Name,
		Email: r.Email,
	}
	m.ID = r.ID
	if r.JobFunction != nil {
		m.JobFunctionID = &r.JobFunction.ID
	}
	for _, g := range r.Groups {
		m.Groups = append(m.Groups, model.StakeholderGroup{Model: model.Model{ID: g.ID}})
	}
	for _, b := range r.BusinessServices {
		m.BusinessServices = append(m.BusinessServices, model.BusinessService{Model: model.Model{ID: b.ID}})
	}
	for _, o := range r.Owns {
		m.Owns = append(m.Owns, model.Application{Model: model.Model{ID: o.ID}})
	}
	for _, c := range r.Contributes {
		m.Contributes = append(m.Contributes, model.Application{Model: model.Model{ID: c.ID}})
	}
	for _, w := range r.MigrationWaves {
		m.MigrationWaves = append(m.MigrationWaves, model.MigrationWave{Model: model.Model{ID: w.ID}})
	}
	return
}
