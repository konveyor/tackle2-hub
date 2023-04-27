package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/konveyor/controller/pkg/logging"
	"github.com/konveyor/tackle2-hub/auth"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	"io"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"time"
)

var Log = logging.WithName("api")

//
// BaseHandler base handler.
type BaseHandler struct {
	// k8s client.
	Client client.Client
}

//
// With configuration.
func (h *BaseHandler) With(client client.Client) {
	h.Client = client
}

//
// DB return db client associated with the context.
func (h *BaseHandler) DB(ctx *gin.Context) (db *gorm.DB) {
	rtx := WithContext(ctx)
	db = rtx.DB.Debug()
	return
}

//
// Paginated returns a paginated & sorted DB client.
func (h *BaseHandler) Paginated(ctx *gin.Context) (db *gorm.DB) {
	p := Page{}
	p.With(ctx)
	db = h.DB(ctx)
	db = p.Paginated(db)
	sort := Sort{}
	sort.With(ctx)
	db = sort.Sorted(db)
	return
}

//
// Sorted returns a sorted DB client.
func (h *BaseHandler) Sorted(ctx *gin.Context) (db *gorm.DB) {
	sort := Sort{}
	sort.With(ctx)
	db = sort.Sorted(h.DB(ctx))
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
	var inspect func(r interface{})
	inspect = func(r interface{}) {
		mt := reflect.TypeOf(r)
		mv := reflect.ValueOf(r)
		if mt.Kind() == reflect.Ptr {
			mt = mt.Elem()
			mv = mv.Elem()
		}
		for i := 0; i < mt.NumField(); i++ {
			ft := mt.Field(i)
			fv := mv.Field(i)
			if !ft.IsExported() {
				continue
			}
			switch fv.Kind() {
			case reflect.Ptr:
				pt := ft.Type.Elem()
				switch pt.Kind() {
				case reflect.Struct, reflect.Slice, reflect.Array:
					continue
				default:
					mp[ft.Name] = fv.Interface()
				}
			case reflect.Struct:
				if ft.Anonymous {
					inspect(fv.Addr().Interface())
				}
			case reflect.Array:
				continue
			case reflect.Slice:
				inst := fv.Interface()
				switch inst.(type) {
				case []byte:
					mp[ft.Name] = fv.Interface()
				}
			default:
				mp[ft.Name] = fv.Interface()
			}
		}
	}
	mp = map[string]interface{}{}
	inspect(m)
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
// Render renders based the Accept: header.
// Opinionated towards json.
func (h *BaseHandler) Render(ctx *gin.Context, code int, r interface{}) {
	ctx.Negotiate(
		code,
		gin.Negotiate{
			Offered: BindMIMEs,
			Data:    r})
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
	ID         uint      `json:"id"`
	CreateUser string    `json:"createUser"`
	UpdateUser string    `json:"updateUser"`
	CreateTime time.Time `json:"createTime"`
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
	mt := reflect.TypeOf(m)
	mv := reflect.ValueOf(m)
	if mv.IsNil() {
		return
	}
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
		mv = mv.Elem()
	}
	for i := 0; i < mt.NumField(); i++ {
		ft := mt.Field(i)
		fv := mv.Field(i)
		switch ft.Name {
		case "Name":
			name = fv.String()
			return
		}
	}
	return
}

//
// Ref represents a FK.
// Contains the PK and (name) natural key.
// The name is read-only.
type Ref struct {
	ID   uint   `json:"id" binding:"required"`
	Name string `json:"name"`
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
type Sort struct {
	Descending bool
	Field      string
}

//
// With context.
func (p *Sort) With(ctx *gin.Context) {
	s := ctx.Query("sort")
	if s == "" {
		return
	}
	part := strings.SplitN(s, ":", 2)
	if len(part) == 2 {
		p.Descending = strings.ToLower(part[0])[0] == 'd'
		p.Field = part[1]
	} else {
		p.Field = part[0]
	}
}

//
// Sorted returns sorted DB.
func (p *Sort) Sorted(in *gorm.DB) (out *gorm.DB) {
	out = in
	if p.Field == "" {
		return
	}
	sort := p.Field
	if p.Descending {
		sort += " DESC"
	}
	out = out.Order(sort)
	return
}
