package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/konveyor/tackle2-hub/auth"
	tasking "github.com/konveyor/tackle2-hub/task"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Response values.
type Response struct {
	Status int
	Body   any
}

// Context custom settings.
type Context struct {
	*gin.Context
	// DB client.
	DB *gorm.DB
	// User
	User string
	// Scope
	Scope struct {
		Granted  []auth.Scope
		Required []string
	}
	// k8s Client
	Client client.Client
	// Response
	Response Response
	// Task manager.
	TaskManager *tasking.Manager
}

// Attach to gin context.
func (r *Context) Attach(ctx *gin.Context) {
	r.Context = ctx
	ctx.Set("RichContext", r)
}

// Detach from gin context
func (r *Context) Detach() {
	delete(r.Context.Keys, "RichContext")
}

// Status sets the values to respond to the request with.
func (r *Context) Status(status int) {
	r.Response = Response{
		Status: status,
		Body:   nil,
	}
}

// Respond sets the values to respond to the request with.
func (r *Context) Respond(status int, body any) {
	r.Response = Response{
		Status: status,
		Body:   body,
	}
}

// RichContext returns a rich context attached to the gin context.
func RichContext(ctx *gin.Context) (rtx *Context) {
	key := "RichContext"
	object, found := ctx.Get(key)
	if !found {
		rtx = &Context{}
		rtx.Attach(ctx)
	} else {
		rtx = object.(*Context)
	}
	rtx.Context = ctx
	return
}

// Transaction handler.
func Transaction(ctx *gin.Context) {
	switch ctx.Request.Method {
	case http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete:
		rtx := RichContext(ctx)
		err := rtx.DB.Transaction(func(tx *gorm.DB) (err error) {
			db := rtx.DB
			rtx.DB = tx
			ctx.Next()
			rtx.DB = db
			if len(ctx.Errors) > 0 {
				err = ctx.Errors[0]
				ctx.Errors = nil
			}
			return
		})
		if err != nil {
			_ = ctx.Error(err)
		}
	}
}

// Render renders the response based on the Accept: header.
// Opinionated towards json.
func Render() gin.HandlerFunc {
	return Renderer{}.Render
}

// Renderer used to render the response body.
type Renderer struct{}

// Render renders the response based on the Accept: header.
// Opinionated towards json.
func (r Renderer) Render(ctx *gin.Context) {
	ctx.Next()
	rtx := RichContext(ctx)
	body := rtx.Response.Body
	if body == nil {
		ctx.Status(rtx.Response.Status)
		return
	}
	switch b := body.(type) {
	case Iterator:
		r.renderIterator(ctx, b)
	default:
		bt := reflect.TypeOf(body)
		bv := reflect.ValueOf(body)
		if bt.Kind() == reflect.Ptr {
			bt = bt.Elem()
			bv = bv.Elem()
		}
		switch bt.Kind() {
		case reflect.Slice:
			r.renderSlice(ctx, bv)
		default:
			ctx.Negotiate(
				rtx.Response.Status,
				gin.Negotiate{
					Offered: BindMIMEs,
					Data:    body})
		}
	}
}

// renderIterator renders an iterator (body).
func (r Renderer) renderIterator(ctx *gin.Context, iter Iterator) {
	file, err := os.CreateTemp("", "render-*")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	encoder, err := NewEncoder(ctx, file)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	encoder.beginList()
	for i := 0; ; i++ {
		next, object := iter.Next()
		if !next {
			break
		}
		if iter.Error != nil {
			_ = ctx.Error(iter.Error)
			return
		}
		encoder.writeItem(0, i, object)
	}
	encoder.endList()
	ctx.File(file.Name())
}

// renderSlice renders a slice (body).
func (r Renderer) renderSlice(ctx *gin.Context, bv reflect.Value) {
	file, err := os.CreateTemp("", "render-*")
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	encoder, err := NewEncoder(ctx, file)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	encoder.beginList()
	for i := 0; i < bv.Len(); i++ {
		v := bv.Index(i)
		object := v.Interface()
		encoder.writeItem(0, i, object)
	}
	encoder.endList()
	ctx.File(file.Name())
}

// NewIterator returns an iterator.
func NewIterator(m any, cursor *Cursor, builder Builder) (iter Iterator) {
	iter = Iterator{
		Model:   m,
		Cursor:  cursor,
		Builder: builder,
	}
	return iter
}

// Builder function used to build rendered resources.
type Builder = func(m []any) (r any, err error)

// Iterator used to iterate a cursor to build and stream the response body.
type Iterator struct {
	Model   any
	Cursor  *Cursor
	Builder Builder
	Error   error
	//
	prev struct {
		id uint
		m  any
	}
}

// Next returns true when the next object has been
// returned by the cursor and built.
func (r *Iterator) Next() (next bool, object any) {
	var batch []any
	for {
		var m any
		if r.prev.id == 0 {
			m, next = r.next()
			if r.Error != nil {
				return
			}
		} else {
			m, next = r.prev.m, true
			r.prev.id = 0
		}
		if !next {
			break
		}
		id := r.id(m)
		if len(batch) == 0 || r.id(batch[0]) == id {
			batch = append(batch, m)
		} else {
			r.prev.id = id
			r.prev.m = m
			break
		}
	}
	next = len(batch) > 0
	if !next {
		return
	}
	object, r.Error = r.Builder(batch)
	return
}

// Close the iterator.
func (r *Iterator) Close() {
	r.Cursor.Close()
}

// model returns a new model.
func (r *Iterator) model() (m any) {
	mt := reflect.TypeOf(r.Model)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
	}
	mv := reflect.New(mt)
	m = mv.Interface()
	return
}

// next returns the next object from the cursor.
func (r *Iterator) next() (m any, next bool) {
	m = r.model()
	next = r.Cursor.Next(m)
	if r.Cursor.Error != nil {
		r.Error = r.Cursor.Error
		return
	}
	return
}

func (r *Iterator) id(m any) (id uint) {
	mt := reflect.TypeOf(m)
	mv := reflect.ValueOf(m)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
		mv = mv.Elem()
	}
	if mt.Kind() != reflect.Struct {
		return
	}
	field := mv.FieldByName("ID")
	if !field.IsValid() {
		return
	}
	v := field.Interface()
	if n, cast := v.(uint); cast {
		id = n
	}
	return
}

// NewEncoder returns an Encoder.
func NewEncoder(ctx *gin.Context, output io.Writer) (encoder Encoder, err error) {
	accepted := ctx.NegotiateFormat(BindMIMEs...)
	switch accepted {
	case "",
		binding.MIMEPOSTForm,
		binding.MIMEJSON:
		encoder = &jsonEncoder{output: output}
	case binding.MIMEYAML:
		encoder = &yamlEncoder{output: output}
	default:
		err = &BadRequestError{"MIME not supported."}
	}

	return
}

// Encoder streamed object encoder.
type Encoder interface {
	begin() Encoder
	end() Encoder
	write(s string) Encoder
	writeStr(s string) Encoder
	field(name string) Encoder
	beginList() Encoder
	endList() Encoder
	writeItem(batch, index int, object any) Encoder
	encode(object any) Encoder
	embed(object any) Encoder
}

type jsonEncoder struct {
	output io.Writer
	fields int
}

func (r *jsonEncoder) begin() Encoder {
	r.write("{")
	return r
}

func (r *jsonEncoder) end() Encoder {
	r.write("}")
	return r
}

func (r *jsonEncoder) write(s string) Encoder {
	_, _ = r.output.Write([]byte(s))
	return r
}

func (r *jsonEncoder) writeStr(s string) Encoder {
	r.write("\"" + s + "\"")
	return r
}

func (r *jsonEncoder) field(s string) Encoder {
	if r.fields > 0 {
		r.write(",")
	}
	r.writeStr(s).write(":")
	r.fields++
	return r
}

func (r *jsonEncoder) beginList() Encoder {
	r.write("[")
	return r
}

func (r *jsonEncoder) endList() Encoder {
	r.write("]")
	return r
}

func (r *jsonEncoder) writeItem(batch, index int, object any) Encoder {
	if batch > 0 || index > 0 {
		r.write(",")
	}
	r.encode(object)
	return r
}

func (r *jsonEncoder) encode(object any) Encoder {
	encoder := json.NewEncoder(r.output)
	_ = encoder.Encode(object)
	return r
}

func (r *jsonEncoder) embed(object any) Encoder {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	_ = encoder.Encode(object)
	s := b.String()
	mp := make(map[string]any)
	err := json.Unmarshal([]byte(s), &mp)
	if err == nil {
		r.fields += len(mp)
		s = s[1 : len(s)-2]
	}
	r.write(s)
	return r
}

type yamlEncoder struct {
	output io.Writer
	fields int
	depth  int
}

func (r *yamlEncoder) begin() Encoder {
	r.write("---\n")
	return r
}

func (r *yamlEncoder) end() Encoder {
	return r
}

func (r *yamlEncoder) write(s string) Encoder {
	s += strings.Repeat("  ", r.depth)
	_, _ = r.output.Write([]byte(s))
	return r
}

func (r *yamlEncoder) writeStr(s string) Encoder {
	r.write("\"" + s + "\"")
	return r
}

func (r *yamlEncoder) field(s string) Encoder {
	if r.fields > 0 {
		r.write("\n")
	}
	r.write(s).write(": ")
	r.fields++
	return r
}

func (r *yamlEncoder) beginList() Encoder {
	r.write("\n")
	r.depth++
	return r
}

func (r *yamlEncoder) endList() Encoder {
	r.depth--
	return r
}

func (r *yamlEncoder) writeItem(batch, index int, object any) Encoder {
	r.encode([]any{object})
	return r
}

func (r *yamlEncoder) encode(object any) Encoder {
	encoder := yaml.NewEncoder(r.output)
	_ = encoder.Encode(object)
	return r
}

func (r *yamlEncoder) embed(object any) Encoder {
	b := new(bytes.Buffer)
	encoder := yaml.NewEncoder(b)
	_ = encoder.Encode(object)
	s := b.String()
	mp := make(map[string]any)
	err := yaml.Unmarshal([]byte(s), &mp)
	if err == nil {
		r.fields += len(mp)
	}
	r.write(s)
	return r
}
