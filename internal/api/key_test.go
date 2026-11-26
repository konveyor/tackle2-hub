package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
)

// keyContext returns a context with a JSON body and the key path parameter.
func keyContext(method, key, body string) (ctx *gin.Context) {
	gin.SetMode(gin.TestMode)
	ctx, _ = gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: Key, Value: key},
	}
	return
}

// keyRejected returns true when the handler reported a bad request.
func keyRejected(ctx *gin.Context) (rejected bool) {
	for _, err := range ctx.Errors {
		if errors.Is(err.Err, &BadRequestError{}) {
			rejected = true
		}
	}
	return
}

func TestValidKey(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	for _, key := range []string{
		"",
		"foobar",
		"ui.target.order",
		"git.insecure.enabled",
		"my-key_1~x",
		"a:b",
	} {
		g.Expect(validKey("Setting key", key)).To(gomega.BeNil(), key)
	}
	for _, key := range []string{
		"foo bar",
		"foo/bar",
		"foo?bar",
		"foo#bar",
		"foo%bar",
		"café",
		"tab\there",
	} {
		err := validKey("Setting key", key)
		g.Expect(err).ToNot(gomega.BeNil(), key)
		g.Expect(errors.Is(err, &BadRequestError{})).To(gomega.BeTrue(), key)
	}
}

func TestSettingCreateKey(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := SettingHandler{}
	ctx := keyContext(http.MethodPost, "", `{"key":"foo bar","value":1}`)
	h.Create(ctx)
	g.Expect(keyRejected(ctx)).To(gomega.BeTrue())
	ctx = keyContext(http.MethodPost, "foo/bar", `1`)
	h.CreateByKey(ctx)
	g.Expect(keyRejected(ctx)).To(gomega.BeTrue())
}

func TestFactKeys(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := ApplicationHandler{}
	// Created with a key that needs encoding.
	ctx := keyContext(http.MethodPost, "", `{"key":"a b","source":"analysis","value":1}`)
	h.FactCreate(ctx)
	g.Expect(keyRejected(ctx)).To(gomega.BeTrue())
	// Created with a source that needs encoding.
	ctx = keyContext(http.MethodPost, "", `{"key":"ab","source":"a/b","value":1}`)
	h.FactCreate(ctx)
	g.Expect(keyRejected(ctx)).To(gomega.BeTrue())
	// Put by key.
	ctx = keyContext(http.MethodPut, "analysis:a b", `1`)
	h.FactPut(ctx)
	g.Expect(keyRejected(ctx)).To(gomega.BeTrue())
	// Replace a source with a name that needs encoding.
	ctx = keyContext(http.MethodPut, "analysis:", `{"a b":1}`)
	h.FactReplace(ctx, FactKey("analysis:"))
	g.Expect(keyRejected(ctx)).To(gomega.BeTrue())
}
