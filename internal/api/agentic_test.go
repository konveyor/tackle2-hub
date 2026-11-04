package api

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	k8sapi "github.com/konveyor/tackle2-hub/internal/k8s/api"
	"github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// agenticContext returns a context with the specified query.
func agenticContext(query string) (ctx *gin.Context) {
	ctx = &gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: query,
			},
			Header: http.Header{},
		},
	}
	return
}

// managedSelector returns the ManagedLabel selector in the options.
func managedSelector(options []k8s.ListOption) (selector k8s.MatchingLabels, found bool) {
	for _, option := range options {
		if m, cast := option.(k8s.MatchingLabels); cast {
			selector = m
			found = true
			return
		}
	}
	return
}

func TestAgenticListAll(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := AgenticHandler{}
	// Omitted.
	all, err := h.listAll(agenticContext(""))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(all).To(gomega.BeFalse())
	// True.
	for _, q := range []string{"all=true", "all=1", "all=True"} {
		all, err = h.listAll(agenticContext(q))
		g.Expect(err).To(gomega.BeNil())
		g.Expect(all).To(gomega.BeTrue())
	}
	// False.
	for _, q := range []string{"all=false", "all=0"} {
		all, err = h.listAll(agenticContext(q))
		g.Expect(err).To(gomega.BeNil())
		g.Expect(all).To(gomega.BeFalse())
	}
	// Not a boolean.
	all, err = h.listAll(agenticContext("all=maybe"))
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(errors.Is(err, &BadRequestError{})).To(gomega.BeTrue())
	g.Expect(all).To(gomega.BeFalse())
}

func TestAgenticListOptions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := AgenticHandler{}
	ns := Settings.Hub.Namespace
	// Omitted: filtered by ManagedLabel (unchanged behavior).
	options, err := h.listOptions(agenticContext(""))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(options).To(gomega.ContainElement(k8s.InNamespace(ns)))
	selector, found := managedSelector(options)
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(selector).To(gomega.Equal(k8s.MatchingLabels{ManagedLabel: "true"}))
	// all=false: filtered by ManagedLabel.
	options, err = h.listOptions(agenticContext("all=false"))
	g.Expect(err).To(gomega.BeNil())
	_, found = managedSelector(options)
	g.Expect(found).To(gomega.BeTrue())
	// all=true: namespace only.
	options, err = h.listOptions(agenticContext("all=true"))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(options).To(gomega.Equal([]k8s.ListOption{k8s.InNamespace(ns)}))
	_, found = managedSelector(options)
	g.Expect(found).To(gomega.BeFalse())
	// Not a boolean: reported.
	_, err = h.listOptions(agenticContext("all=maybe"))
	g.Expect(err).ToNot(gomega.BeNil())
}

// agenticClient returns a client with the specified objects.
func agenticClient(g *gomega.WithT, objects ...k8s.Object) (client k8s.Client) {
	scheme := runtime.NewScheme()
	err := k8sapi.AddToScheme(scheme)
	g.Expect(err).To(gomega.BeNil())
	builder := fake.NewClientBuilder().WithScheme(scheme)
	for _, object := range objects {
		builder = builder.WithObjects(object)
	}
	client = builder.Build()
	return
}

// agenticNames returns the names in the (list) response.
func agenticNames(ctx *gin.Context) (names []string) {
	names = []string{}
	body := RichContext(ctx).Response.Body
	for _, agent := range body.([]Agent) {
		names = append(names, agent.Name)
	}
	return
}

// agenticListed returns the names listed by AgentList.
func agenticListed(g *gomega.WithT, client k8s.Client, query string) (names []string, status int) {
	h := AgenticHandler{}
	ctx := agenticContext(query)
	rtx := &Context{Client: client}
	rtx.Attach(ctx)
	h.AgentList(ctx)
	status = rtx.Response.Status
	if len(ctx.Errors) == 0 {
		names = agenticNames(ctx)
	}
	return
}

func TestAgentListManaged(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	ns := Settings.Hub.Namespace
	// Created through the API.
	managed := &Agent{
		ObjectMeta: v1.ObjectMeta{
			Namespace: ns,
			Name:      "managed",
			Labels: map[string]string{
				ManagedLabel: "true",
			},
		},
	}
	// Created with (for example) kubectl apply.
	applied := &Agent{
		ObjectMeta: v1.ObjectMeta{
			Namespace: ns,
			Name:      "applied",
		},
	}
	client := agenticClient(g, managed, applied)
	// Omitted: managed only (unchanged behavior).
	names, status := agenticListed(g, client, "")
	g.Expect(status).To(gomega.Equal(http.StatusOK))
	g.Expect(names).To(gomega.Equal([]string{"managed"}))
	// all=false: managed only.
	names, status = agenticListed(g, client, "all=false")
	g.Expect(status).To(gomega.Equal(http.StatusOK))
	g.Expect(names).To(gomega.Equal([]string{"managed"}))
	// all=true: both.
	names, status = agenticListed(g, client, "all=true")
	g.Expect(status).To(gomega.Equal(http.StatusOK))
	g.Expect(names).To(gomega.ConsistOf("managed", "applied"))
	// Not a boolean: reported, nothing listed.
	names, _ = agenticListed(g, client, "all=maybe")
	g.Expect(names).To(gomega.BeNil())
}
