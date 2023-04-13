package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

// Routes
const (
	TicketsRoot = "/tickets"
	TicketRoot  = "/tickets" + "/:" + ID
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
	routeGroup.GET(TicketsRoot, h.List)
	routeGroup.GET(TicketsRoot+"/", h.List)
	routeGroup.POST(TicketsRoot, h.Create)
	routeGroup.GET(TicketRoot, h.Get)
	routeGroup.DELETE(TicketRoot, h.Delete)
}

// Get godoc
// @summary Get a ticket by ID.
// @description Get a ticket by ID.
// @tags tickets
// @produce json
// @success 200 {object} api.Ticket
// @router /tickets/{id} [get]
// @param id path string true "Ticket ID"
func (h TicketHandler) Get(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Ticket{}
	db := h.preLoad(h.DB(ctx), clause.Associations)
	result := db.First(m, id)
	if result.Error != nil {
		_ = ctx.Error(result.Error)
		return
	}

	resource := Ticket{}
	resource.With(m)
	h.Render(ctx, http.StatusOK, resource)
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
	db := h.preLoad(h.Paginated(ctx), clause.Associations)
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

	h.Render(ctx, http.StatusOK, resources)
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

	h.Render(ctx, http.StatusCreated, r)
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

	ctx.Status(http.StatusNoContent)
}

// Ticket API Resource
type Ticket struct {
	Resource
	Kind        string    `json:"kind" binding:"required"`
	Reference   string    `json:"reference"`
	Link        string    `json:"link"`
	Parent      string    `json:"parent" binding:"required"`
	Error       bool      `json:"error"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
	Fields      Fields    `json:"fields"`
	Application Ref       `json:"application" binding:"required"`
	Tracker     Ref       `json:"tracker" binding:"required"`
}

// With updates the resource with the model.
func (r *Ticket) With(m *model.Ticket) {
	r.Resource.With(&m.Model)
	r.Kind = m.Kind
	r.Reference = m.Reference
	r.Parent = m.Parent
	r.Link = m.Link
	r.Error = m.Error
	r.Message = m.Message
	r.Status = m.Status
	r.LastUpdated = m.LastUpdated
	r.Application = r.ref(m.ApplicationID, m.Application)
	r.Tracker = r.ref(m.TrackerID, m.Tracker)
	_ = json.Unmarshal(m.Fields, &r.Fields)
}

// Model builds a model.
func (r *Ticket) Model() (m *model.Ticket) {
	m = &model.Ticket{
		Kind:          r.Kind,
		Parent:        r.Parent,
		ApplicationID: r.Application.ID,
		TrackerID:     r.Tracker.ID,
	}
	if r.Fields == nil {
		r.Fields = Fields{}
	}
	m.Fields, _ = json.Marshal(r.Fields)
	m.ID = r.ID

	return
}

type Fields map[string]interface{}
