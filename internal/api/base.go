package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/api/association"
	"github.com/konveyor/tackle2-hub/internal/api/jsd"
	"github.com/konveyor/tackle2-hub/internal/api/sort"
	"github.com/konveyor/tackle2-hub/internal/auth"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Log = logr.New("api", Settings.Log.Web)

const (
	MaxPage  = 500
	MaxCount = 50000
)

// BaseHandler base handler.
type BaseHandler struct{}

// DB return db client associated with the context.
func (h *BaseHandler) DB(ctx *gin.Context) (db *gorm.DB) {
	rtx := RichContext(ctx)
	if Log.V(1).Enabled() {
		db = rtx.DB.Debug()
	} else {
		db = rtx.DB
	}
	return
}

// Association returns an association (manager).
func (h *BaseHandler) Association(ctx *gin.Context, name string) *association.Association {
	return association.New(h.DB(ctx), name)
}

// Client returns k8s client from the context.
func (h *BaseHandler) Client(ctx *gin.Context) (client client.Client) {
	rtx := RichContext(ctx)
	client = rtx.Client
	return
}

// WithCount report count.
// Sets the X-Total header for pagination.
// Returns an error when count exceeds the limited and
// is not constrained by pagination.
func (h *BaseHandler) WithCount(ctx *gin.Context, count int64) (err error) {
	n := int(count)
	p := Page{}
	p.With(ctx)
	if n > MaxPage {
		if p.Limit == 0 || p.Limit > MaxPage {
			err = &BadRequestError{
				Reason: fmt.Sprintf(
					"Found=%d, ?Limit <= %d required.",
					n,
					MaxPage)}
			return
		}
	}
	s := strconv.Itoa(n)
	if n > MaxCount {
		s = ">" + strconv.Itoa(MaxCount)
	}
	mp := ctx.Writer.Header()
	mp[Total] = []string{s}
	return
}

// preLoad update DB to pre-load fields.
func (h *BaseHandler) preLoad(db *gorm.DB, fields ...string) (tx *gorm.DB) {
	tx = db
	for _, f := range fields {
		tx = tx.Preload(f)
	}

	return
}

// pk returns the PK (ID) parameter.
func (h *BaseHandler) pk(ctx *gin.Context) (id uint) {
	s := ctx.Param(ID)
	n, _ := strconv.Atoi(s)
	id = uint(n)
	return
}

// CurrentUser gets username from Keycloak auth token.
func (h *BaseHandler) CurrentUser(ctx *gin.Context) (user string) {
	rtx := RichContext(ctx)
	user = rtx.User
	if user == "" {
		Log.Info("Failed to get current user.")
	}

	return
}

// HasScope determines if the token has the specified scope.
func (h *BaseHandler) HasScope(ctx *gin.Context, scope string) (b bool) {
	in := auth.BaseScope{}
	in.With(scope)
	rtx := RichContext(ctx)
	for _, s := range rtx.Scope.Granted {
		b = s.Match(in.Resource, in.Method)
		if b {
			return
		}
	}
	return
}

// Encrypt the model.
func (h *BaseHandler) Encrypt(m any) (err error) {
	err = secret.Encrypt(m)
	return
}

// Decrypt the model.
// When:
//   - decrypted parameter true.
//   - user has required scope.
func (h *BaseHandler) Decrypt(ctx *gin.Context, m any) (err error) {
	q := ctx.Query(Decrypted)
	requested, _ := strconv.ParseBool(q)
	if !requested {
		return
	}
	rtx := RichContext(ctx)
	for _, scope := range rtx.Scope.Required {
		scope += ":" + MethodDecrypt
		if h.HasScope(ctx, scope) {
			err = secret.Decrypt(m)
			return
		}
	}
	err = &Forbidden{
		Reason: ":decrypt (scope) required.",
	}
	return
}

// Bind based on Content-Type header.
// Opinionated towards json.
func (h *BaseHandler) Bind(ctx *gin.Context, r any) (err error) {
	switch ctx.ContentType() {
	case "",
		binding.MIMEPOSTForm,
		binding.MIMEJSON:
		err = h.BindJSON(ctx, r)
	case binding.MIMEYAML:
		err = h.BindYAML(ctx, r)
	default:
		err = &BadRequestError{Reason: "Bind: MIME not supported."}
	}
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
	}
	return
}

// BindJSON attempts to bind a request body to a struct, assuming that the body is JSON.
// Binding is strict: unknown fields in the input will cause binding to fail.
func (h *BaseHandler) BindJSON(ctx *gin.Context, r any) (err error) {
	if ctx.Request == nil || ctx.Request.Body == nil {
		err = errors.New("invalid request")
		return
	}
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(r)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = h.Validate(ctx, r)
	return
}

// BindYAML attempts to bind a request body to a struct, assuming that the body is YAML.
// Binding is strict: unknown fields in the input will cause binding to fail.
func (h *BaseHandler) BindYAML(ctx *gin.Context, r any) (err error) {
	if ctx.Request == nil || ctx.Request.Body == nil {
		err = errors.New("invalid request")
		return
	}
	decoder := yaml.NewDecoder(ctx.Request.Body)
	decoder.SetStrict(true)
	err = decoder.Decode(r)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = h.Validate(ctx, r)
	return
}

// Validate that the struct field values obey the binding field tags.
func (h *BaseHandler) Validate(ctx *gin.Context, r any) (err error) {
	if binding.Validator == nil {
		return
	}
	err = binding.Validator.ValidateStruct(r)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	rtx := RichContext(ctx)
	jsdValidator := jsd.New(rtx.Client)
	err = jsdValidator.Validate(r)
	return
}

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
		err = &BadRequestError{Reason: "Bind: MIME not supported."}
	}
	if err != nil {
		err = &BadRequestError{Reason: err.Error()}
	}
	return
}

// Status sets the status code.
func (h *BaseHandler) Status(ctx *gin.Context, code int) {
	rtx := RichContext(ctx)
	rtx.Status(code)
}

// Respond sets the response.
func (h *BaseHandler) Respond(ctx *gin.Context, code int, r any) {
	rtx := RichContext(ctx)
	rtx.Respond(code, r)
}

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

// Attachment sets the Content-Disposition header.
func (h *BaseHandler) Attachment(ctx *gin.Context, name string) {
	attachment := fmt.Sprintf("attachment; filename=\"%s\"", name)
	ctx.Writer.Header().Set(
		"Content-Disposition",
		attachment)
}

// Page provides pagination.
type Page struct {
	Offset int
	Limit  int
}

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

// Sort provides sorting.
type Sort = sort.Sort

// Decoder binding decoder.
type Decoder interface {
	Decode(r any) (err error)
}

// Cursor Paginated rows iterator.
type Cursor struct {
	Page
	DB    *gorm.DB
	Rows  *sql.Rows
	Index int64
	Error error
}

// Next returns true when has next row.
func (r *Cursor) Next(m any) (next bool) {
	defer func() {
		p := recover()
		if p != nil {
			switch p.(type) {
			case error:
				r.Error = p.(error)
			default:
				r.Error = errors.New(fmt.Sprint(p))
			}
		}
	}()
	if r.Error != nil {
		next = true
		return
	}
	next = r.Rows.Next()
	if next {
		r.Index++
	} else {
		return
	}
	if r.pageLimited() {
		for r.Rows.Next() {
			r.Index++
			if r.Index > MaxCount {
				break
			}
		}
		next = false
		r.Close()
		return
	}
	r.Error = r.DB.ScanRows(r.Rows, m)
	return
}

// With configures the cursor.
func (r *Cursor) With(db *gorm.DB, p Page) {
	r.DB = db.Offset(p.Offset)
	r.Rows, r.Error = r.DB.Rows()
	r.Index = int64(0)
	r.Page = p
}

// Count returns the count adjusted for offset.
func (r *Cursor) Count() (n int64) {
	n = int64(r.Offset) + r.Index
	return n
}

// Close the cursor.
func (r *Cursor) Close() {
	if r.Rows != nil {
		_ = r.Rows.Close()
	}
}

// pageLimited returns true when page Limit defined and exceeded.
func (r *Cursor) pageLimited() (b bool) {
	if r.Limit < 1 {
		return
	}
	b = r.Index > int64(r.Limit)
	return
}
