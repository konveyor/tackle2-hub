package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jortel/go-utils/logr"
	reflect "github.com/konveyor/tackle2-hub/api/reflect"
	"github.com/konveyor/tackle2-hub/api/sort"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"io"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"time"
)

var Log = logr.WithName("api")

//
// BaseHandler base handler.
type BaseHandler struct{}

//
// DB return db client associated with the context.
func (h *BaseHandler) DB(ctx *gin.Context) (db *gorm.DB) {
	rtx := WithContext(ctx)
	db = rtx.DB.Debug()
	return
}

//
// Client returns k8s client from the context.
func (h *BaseHandler) Client(ctx *gin.Context) (client client.Client) {
	rtx := WithContext(ctx)
	client = rtx.Client
	return
}

//
// WithCount report count.
// Sets the X-Total header for pagination.
// Returns an error when count exceeds the limited and
// is not constrained by pagination.
func (h *BaseHandler) WithCount(ctx *gin.Context, count int64) (err error) {
	n := int(count)
	max := 500
	p := Page{}
	p.With(ctx)
	if n > max {
		if p.Limit == 0 || p.Limit > max {
			err = &BadRequestError{
				fmt.Sprintf(
					"Found=%d, ?Limit <= %d required.",
					n,
					max)}
			return
		}
	}
	mp := ctx.Writer.Header()
	mp[Total] = []string{
		strconv.Itoa(int(count)),
	}
	return
}

//
// Paginated returns a paginated and sorted DB client.
func (h *BaseHandler) paginated(ctx *gin.Context, sort Sort, in *gorm.DB) (db *gorm.DB) {
	p := Page{}
	p.With(ctx)
	db = p.Paginated(in)
	db = sort.Sorted(db)
	return
}

//
// preLoad update DB to pre-load fields.
func (h *BaseHandler) preLoad(db *gorm.DB, fields ...string) (tx *gorm.DB) {
	tx = db
	for _, f := range fields {
		tx = tx.Preload(f)
	}

	return
}

//
// fields builds a map of fields.
func (h *BaseHandler) fields(m interface{}) (mp map[string]interface{}) {
	mp = reflect.Fields(m)
	return
}

//
// pk returns the PK (ID) parameter.
func (h *BaseHandler) pk(ctx *gin.Context) (id uint) {
	s := ctx.Param(ID)
	n, _ := strconv.Atoi(s)
	id = uint(n)
	return
}

//
// modBody updates the body using the `mod` function.
//   1. read the body.
//   2. mod()
//   3. write body.
func (h *BaseHandler) modBody(
	ctx *gin.Context,
	r interface{},
	mod func(bool) error) (err error) {
	//
	withBody := false
	if ctx.Request.ContentLength > 0 {
		withBody = true
		err = h.Bind(ctx, r)
		if err != nil {
			return
		}
	}
	err = mod(withBody)
	if err != nil {
		return
	}
	b, _ := json.Marshal(r)
	bfr := bytes.NewBuffer(b)
	ctx.Request.Body = io.NopCloser(bfr)
	return
}

//
// CurrentUser gets username from Keycloak auth token.
func (h *BaseHandler) CurrentUser(ctx *gin.Context) (user string) {
	rtx := WithContext(ctx)
	user = rtx.User
	if user == "" {
		Log.Info("Failed to get current user.")
	}

	return
}

//
// HasScope determines if the token has the specified scope.
func (h *BaseHandler) HasScope(ctx *gin.Context, scope string) (b bool) {
	in := auth.BaseScope{}
	in.With(scope)
	rtx := WithContext(ctx)
	for _, s := range rtx.Scopes {
		b = s.Match(in.Resource, in.Method)
		if b {
			return
		}
	}
	return
}

//
// Bind based on Content-Type header.
// Opinionated towards json.
func (h *BaseHandler) Bind(ctx *gin.Context, r interface{}) (err error) {
	switch ctx.ContentType() {
	case "",
		binding.MIMEPOSTForm,
		binding.MIMEJSON:
		err = ctx.BindJSON(r)
	case binding.MIMEYAML:
		err = ctx.BindYAML(r)
	default:
		err = &BadRequestError{"Bind: MIME not supported."}
	}
	if err != nil {
		err = &BadRequestError{err.Error()}
	}
	return
}

//
// Decoder returns a decoder based on encoding.
// Opinionated towards json.
func (h *BaseHandler) Decoder(ctx *gin.Context, encoding string, r io.Reader) (d Decoder, err error) {
	if r == nil {
		r = ctx.Request.Body
	}
	switch encoding {
	case "",
		binding.MIMEPOSTForm,
		binding.MIMEMultipartPOSTForm,
		binding.MIMEJSON:
		d = json.NewDecoder(r)
	case binding.MIMEYAML:
		d = yaml.NewDecoder(r)
	default:
		err = &BadRequestError{"Bind: MIME not supported."}
	}
	if err != nil {
		err = &BadRequestError{err.Error()}
	}
	return
}

//
// Status sets the status code.
func (h *BaseHandler) Status(ctx *gin.Context, code int) {
	rtx := WithContext(ctx)
	rtx.Status(code)
}

//
// Respond sets the response.
func (h *BaseHandler) Respond(ctx *gin.Context, code int, r interface{}) {
	rtx := WithContext(ctx)
	rtx.Respond(code, r)
}

//
// Accepted determines if the mime is accepted.
// Wildcards ignored.
func (h *BaseHandler) Accepted(ctx *gin.Context, mimes ...string) (b bool) {
	accept := ctx.GetHeader(Accept)
	for _, accepted := range strings.Split(accept, ",") {
		accepted = strings.TrimSpace(accepted)
		accepted = strings.Split(accepted, ";")[0]
		for _, wanted := range mimes {
			if accepted == wanted {
				b = true
				return
			}
		}
	}
	return
}

//
// REST resource.
type Resource struct {
	ID         uint      `json:"id,omitempty" yaml:",omitempty"`
	CreateUser string    `json:"createUser" yaml:",omitempty"`
	UpdateUser string    `json:"updateUser" yaml:",omitempty"`
	CreateTime time.Time `json:"createTime" yaml:",omitempty"`
}

//
// With updates the resource with the model.
func (r *Resource) With(m *model.Model) {
	r.ID = m.ID
	r.CreateUser = m.CreateUser
	r.UpdateUser = m.UpdateUser
	r.CreateTime = m.CreateTime
}

//
// ref with id and named model.
func (r *Resource) ref(id uint, m interface{}) (ref Ref) {
	ref.ID = id
	ref.Name = r.nameOf(m)
	return
}

//
// refPtr with id and named model.
func (r *Resource) refPtr(id *uint, m interface{}) (ref *Ref) {
	if id == nil {
		return
	}
	ref = &Ref{}
	ref.ID = *id
	ref.Name = r.nameOf(m)
	return
}

//
// idPtr extracts ref ID.
func (r *Resource) idPtr(ref *Ref) (id *uint) {
	if ref != nil {
		id = &ref.ID
	}
	return
}

//
// nameOf model.
func (r *Resource) nameOf(m interface{}) (name string) {
	name = reflect.NameOf(m)
	return
}

//
// Ref represents a FK.
// Contains the PK and (name) natural key.
// The name is optional and read-only.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name,omitempty"`
}

//
// With id and named model.
func (r *Ref) With(id uint, name string) {
	r.ID = id
	r.Name = name
}

//
// TagRef represents a reference to a Tag.
// Contains the tag ID, name, tag source.
type TagRef struct {
	ID     uint   `json:"id" binding:"required"`
	Name   string `json:"name"`
	Source string `json:"source"`
}

//
// With id and named model.
func (r *TagRef) With(id uint, name string, source string) {
	r.ID = id
	r.Name = name
	r.Source = source
}

//
// Page provides pagination.
type Page struct {
	Offset int
	Limit  int
}

//
// With context.
func (p *Page) With(ctx *gin.Context) {
	s := ctx.Query("offset")
	if s != "" {
		p.Offset, _ = strconv.Atoi(s)
	}
	s = ctx.Query("limit")
	if s != "" {
		p.Limit, _ = strconv.Atoi(s)
	}
	return
}

//
// Paginated returns a paginated DB.
func (p *Page) Paginated(in *gorm.DB) (out *gorm.DB) {
	out = in
	if p.Offset > 0 {
		out = out.Offset(p.Offset)
	}
	if p.Limit > 0 {
		out = out.Limit(p.Limit)
	}
	return
}

//
// Sort provides sorting.
type Sort = sort.Sort

//
// Decoder binding decoder.
type Decoder interface {
	Decode(r interface{}) (err error)
}
