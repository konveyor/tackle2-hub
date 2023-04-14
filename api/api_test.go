package api

import (
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestAccepted(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := BaseHandler{}
	ctx := &gin.Context{
		Request: &http.Request{
			Header: http.Header{},
		},
	}
	// plain
	ctx.Request.Header[Accept] = []string{"a/b"}
	g.Expect(h.Accepted(ctx, "a/b")).To(gomega.BeTrue())
	ctx.Request.Header[Accept] = []string{"a/b"}
	g.Expect(h.Accepted(ctx, "x/y")).To(gomega.BeFalse())
	// Empty.
	ctx.Request.Header[Accept] = []string{""}
	g.Expect(h.Accepted(ctx, "x/y")).To(gomega.BeFalse())
	ctx.Request.Header[Accept] = []string{"a/b"}
	g.Expect(h.Accepted(ctx, "")).To(gomega.BeFalse())
	// Multiple with spaces.
	ctx.Request.Header[Accept] = []string{"x/y, a/b, c/d"}
	g.Expect(h.Accepted(ctx, "a/b")).To(gomega.BeTrue())
	// Multiple and parameters.
	ctx.Request.Header[Accept] = []string{"x/y,a/b;q=1.0"}
	g.Expect(h.Accepted(ctx, "a/b")).To(gomega.BeTrue())
}
