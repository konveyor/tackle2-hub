package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

//
// Routes
const (
	BatchRoot        = "/batch"
	BatchTicketsRoot = BatchRoot + TicketsRoot
	BatchTagsRoot    = BatchRoot + TagsRoot
)

//
// BatchHandler handles batch resource creation routes.
type BatchHandler struct {
	BaseHandler
}

//
// AddRoutes adds routes.
func (h BatchHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.POST(BatchTicketsRoot, Required("tickets"), Transaction, h.TicketsCreate)
	routeGroup.POST(BatchTagsRoot, Required("tags"), Transaction, h.TagsCreate)
}

// TicketsCreate godoc
// @summary Batch-create Tickets.
// @description Batch-create Tickets.
// @tags batch, tickets
// @produce json
// @success 200 {object} []api.Ticket
// @router /batch/tickets [post]
// @param tickets body []api.Ticket true "Tickets data"
func (h BatchHandler) TicketsCreate(ctx *gin.Context) {
	handler := TicketHandler{}
	h.create(ctx, handler.Create)
}

// TagsCreate godoc
// @summary Batch-create Tags.
// @description Batch-create Tags.
// @tags batch, tags
// @produce json
// @success 200 {object} []api.Tag
// @router /batch/tags [post]
// @param tags body []api.Tag true "Tags data"
func (h BatchHandler) TagsCreate(ctx *gin.Context) {
	handler := TagHandler{}
	h.create(ctx, handler.Create)
}

func (h BatchHandler) create(ctx *gin.Context, create gin.HandlerFunc) {
	var resources []interface{}
	err := h.Bind(ctx, &resources)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	rtx := WithContext(ctx)
	bErr := BatchError{Message: "Create failed."}
	for i := range resources {
		b, _ := json.Marshal(resources[i])
		bfr := bytes.NewBuffer(b)
		ctx.Request.Body = io.NopCloser(bfr)
		create(ctx)
		if len(ctx.Errors) > 0 {
			err = ctx.Errors[0]
			bErr.Items = append(bErr.Items, BatchErrorItem{
				Error:    err,
				Resource: resources[i],
			})
			ctx.Errors = nil
			continue
		}
		resources[i] = rtx.Response.Body
	}
	if len(bErr.Items) == 0 {
		h.Respond(ctx, http.StatusCreated, resources)
	} else {
		_ = ctx.Error(bErr)
	}
}
