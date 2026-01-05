package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
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

func TestFactKey(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	key := FactKey("key")
	g.Expect(key.Source()).To(gomega.Equal(""))
	g.Expect(key.Name()).To(gomega.Equal("key"))

	key = FactKey("test:key")
	g.Expect(key.Source()).To(gomega.Equal("test"))
	g.Expect(key.Name()).To(gomega.Equal("key"))

	key = FactKey(":key")
	g.Expect(key.Source()).To(gomega.Equal(""))
	g.Expect(key.Name()).To(gomega.Equal("key"))

	key = FactKey("test:")
	g.Expect(key.Source()).To(gomega.Equal("test"))
	g.Expect(key.Name()).To(gomega.Equal(""))

	key = FactKey("")
	key.Qualify("test")
	g.Expect(key.Source()).To(gomega.Equal("test"))
	g.Expect(key.Name()).To(gomega.Equal(""))

	key = FactKey("key")
	key.Qualify("test")
	g.Expect(key.Source()).To(gomega.Equal("test"))
	g.Expect(key.Name()).To(gomega.Equal("key"))

	key = FactKey("source:key")
	key.Qualify("test")
	g.Expect(key.Source()).To(gomega.Equal("test"))
	g.Expect(key.Name()).To(gomega.Equal("key"))

	key = FactKey("source:")
	key.Qualify("test")
	g.Expect(key.Source()).To(gomega.Equal("test"))
	g.Expect(key.Name()).To(gomega.Equal(""))
}

func TestEncoder(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Thing struct {
		Name string
		Age  int
		List []string
	}
	thing := &Thing{
		Name: "elmer",
		Age:  1,
		List: []string{"a", "b", "c", "d"},
	}

	type Test struct {
		encoder Encoder
		decoder func([]byte, any) error
	}

	cases := []Test{
		{
			encoder: &jsonEncoder{},
			decoder: func(b []byte, object any) error {
				return json.Unmarshal(b, object)
			}},
		{
			encoder: &yamlEncoder{},
			decoder: func(b []byte, object any) error {
				return yaml.Unmarshal(b, object)
			}},
	}

	for _, tc := range cases {
		en := tc.encoder
		b := &bytes.Buffer{}
		if x, cast := en.(*jsonEncoder); cast {
			x.output = b
		}
		if x, cast := en.(*yamlEncoder); cast {
			x.output = b
		}
		en.begin()
		en.embed(thing)
		en.field("field")
		en.writeStr("value")
		en.field("things")
		en.beginList()
		en.writeItem(0, 0, thing)
		en.endList()
		en.field("items")
		en.beginList()
		en.writeItem(0, 0, "item-0")
		en.writeItem(0, 1, "item-1")
		en.writeItem(0, 2, "item-2")
		en.endList()
		en.node("thing", thing)
		en.end()
		s := b.String()
		//println(s)

		err := en.error()
		g.Expect(err).To(gomega.BeNil())

		decoded := map[string]any{}
		err = tc.decoder([]byte(s), &decoded)
		g.Expect(err).To(gomega.BeNil())
	}
}

func TestEncoderError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	for _, en := range []Encoder{&jsonEncoder{}, &yamlEncoder{}} {
		err := fmt.Errorf("")
		b := &_errorWriter{err: err}
		if x, cast := en.(*jsonEncoder); cast {
			x.output = b
		}
		if x, cast := en.(*yamlEncoder); cast {
			x.output = b
		}
		en.begin()
		en.field("root")
		en.write("value")
		en.embed(nil)
		en.end()

		g.Expect(err).To(gomega.Equal(en.error()))
	}
}

func TestManifestReader(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	b := make([]byte, 1024)

	//
	// Test marker content.

	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, _ = f.WriteString(BeginMainMarker)
	_, _ = f.WriteString("\n")
	_, _ = f.WriteString("Main-ONE\n")
	_, _ = f.WriteString("Main-TWO\n")
	_, _ = f.WriteString(EndMainMarker)
	_, _ = f.WriteString("\n")

	_, _ = f.WriteString(BeginInsightsMarker)
	_, _ = f.WriteString("\r\n")
	_, _ = f.WriteString("Insight-ONE\n")
	_, _ = f.WriteString("Insight-TWO\n")
	_, _ = f.WriteString(EndInsightsMarker)
	_, _ = f.WriteString("\r\n")

	_, _ = f.WriteString(BeginDepsMarker)
	_, _ = f.WriteString("\n")
	_, _ = f.WriteString("Dep-ONE\n")
	_, _ = f.WriteString("Dep-TWO\n")
	_, _ = f.WriteString(EndDepsMarker)
	_, _ = f.WriteString("\n")
	_ = f.Close()

	r := ManifestReader{}
	err = r.Open(f.Name(), BeginMainMarker, EndMainMarker)
	g.Expect(err).To(gomega.BeNil())
	n, err := r.Read(b)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(19).To(gomega.Equal(n))
	s := string(b[:n])
	g.Expect("\nMain-ONE\nMain-TWO\n").To(gomega.Equal(s))
	_ = r.Close()

	r = ManifestReader{}
	err = r.Open(f.Name(), BeginInsightsMarker, EndInsightsMarker)
	g.Expect(err).To(gomega.BeNil())
	n, err = r.Read(b)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(26).To(gomega.Equal(n))
	s = string(b[:n])
	g.Expect("\r\nInsight-ONE\nInsight-TWO\n").To(gomega.Equal(s))
	_ = r.Close()

	r = ManifestReader{}
	err = r.Open(f.Name(), BeginDepsMarker, EndDepsMarker)
	g.Expect(err).To(gomega.BeNil())
	n, err = r.Read(b)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(17).To(gomega.Equal(n))
	s = string(b[:n])
	g.Expect("\nDep-ONE\nDep-TWO\n").To(gomega.Equal(s))
	_ = r.Close()

	//
	// Test marker not found.

	r = ManifestReader{}
	err = r.Open(f.Name(), "XX", EndInsightsMarker)
	g.Expect(err).ToNot(gomega.BeNil())
	err = r.Open(f.Name(), BeginInsightsMarker, "XX")
	g.Expect(err).ToNot(gomega.BeNil())
	err = r.Open(f.Name(), EndDepsMarker, BeginDepsMarker)
	g.Expect(err).ToNot(gomega.BeNil())

	//
	// Test empty marker sections.

	f2, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		_ = f2.Close()
		_ = os.Remove(f2.Name())
	}()
	_, _ = f2.WriteString(BeginMainMarker)
	_, _ = f2.WriteString("\n")
	_, _ = f2.WriteString(EndMainMarker)
	_, _ = f2.WriteString("\n")
	_ = f2.Close()
	r = ManifestReader{}
	err = r.Open(f2.Name(), BeginMainMarker, EndMainMarker)
	g.Expect(err).To(gomega.BeNil())
	n, err = r.Read(b)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(1).To(gomega.Equal(n))
}

type _errorWriter struct {
	err error
}

func (f *_errorWriter) Write(p []byte) (int, error) {
	return 0, f.err
}
