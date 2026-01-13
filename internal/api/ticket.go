package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// Params.
const (
	TrackerId = "tracker"
)

// TicketHandler handles ticket routes.
type TicketHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h TicketHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("tickets"))
	routeGroup.GET(api.TicketsRoute, h.List)
	routeGroup.GET(api.TicketsRoute+"/", h.List)
	routeGroup.POST(api.TicketsRoute, h.Create)
	routeGroup.GET(api.TicketRoute, h.Get)
	routeGroup.DELETE(api.TicketRoute, h.Delete)
}

// Get godoc
// @summary Get a ticket by ID.
// @description Get a ticket by ID.
// @tags tickets
// @produce json
// @success 200 {object} api.Ticket
// @router /tickets/{id} [get]
// @param id path int true "Ticket ID"
func (h TicketHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Ticket{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	r := Ticket{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all tickets.
// @description List all tickets.
// @tags tickets
// @produce json
// @success 200 {object} []api.Ticket
// @router /tickets [get]
func (h TicketHandler) List(ctx *gin.Context) {
	var list []model.Ticket
	appId := ctx.Query(AppId)
	trackerId := ctx.Query(TrackerId)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	if appId != "" {
		db = db.Where("ApplicationID = ?", appId)
	}
	if trackerId != "" {
		db = db.Where("TrackerID = ?", trackerId)
	}
	result := db.Find(&list)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}
	resources := []Ticket{}
	for i := range list {
		r := Ticket{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a ticket.
// @description Create a ticket.
// @tags tickets
// @accept json
// @produce json
// @success 201 {object} api.Ticket
// @router /tickets [post]
// @param ticket body api.Ticket true "Ticket data"
func (h TicketHandler) Create(ctx *gin.Context) {
	r := &Ticket{}
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

	h.Respond(ctx, http.StatusCreated, r)
}

// Delete godoc
// @summary Delete a ticket.
// @description Delete a ticket.
// @tags tickets
// @success 204
// @router /tickets/{id} [delete]
// @param id path int true "Ticket id"
func (h TicketHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Ticket{}
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

// Ticket API Resource
type Ticket = resource.Ticket
