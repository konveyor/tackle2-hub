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
	"k8s.io/apimachinery/pkg/labels"
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

// agenticSelector returns the label selector in the options.
func agenticSelector(options []k8s.ListOption) (selector labels.Selector, found bool) {
	for _, option := range options {
		if s, cast := option.(k8s.MatchingLabelsSelector); cast {
			selector = s.Selector
			found = true
			return
		}
	}
	return
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
func agenticListed(client k8s.Client, query string) (names []string, status int) {
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

func TestAgenticListOptions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := AgenticHandler{}
	ns := Settings.Hub.Namespace
	managed := labels.Set{ManagedLabel: "true"}
	unmanaged := labels.Set{}
	// Omitted: the namespace, unselected.
	options, err := h.listOptions(agenticContext(""))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(options).To(gomega.Equal([]k8s.ListOption{k8s.InNamespace(ns)}))
	// managed:true.
	options, err = h.listOptions(agenticContext("filter=managed:true"))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(options).To(gomega.ContainElement(k8s.InNamespace(ns)))
	selector, found := agenticSelector(options)
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(selector.Matches(managed)).To(gomega.BeTrue())
	g.Expect(selector.Matches(unmanaged)).To(gomega.BeFalse())
	// managed:false selects resources without the label.
	options, err = h.listOptions(agenticContext("filter=managed:false"))
	g.Expect(err).To(gomega.BeNil())
	selector, found = agenticSelector(options)
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(selector.Matches(managed)).To(gomega.BeFalse())
	g.Expect(selector.Matches(unmanaged)).To(gomega.BeTrue())
	// Not a boolean.
	_, err = h.listOptions(agenticContext("filter=managed:maybe"))
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(errors.Is(err, &BadRequestError{})).To(gomega.BeTrue())
	// Not a single value.
	_, err = h.listOptions(agenticContext("filter=managed:(true|false)"))
	g.Expect(err).ToNot(gomega.BeNil())
	// Not a supported field.
	_, err = h.listOptions(agenticContext("filter=name:elmer"))
	g.Expect(err).ToNot(gomega.BeNil())
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
	// Omitted: both.
	names, status := agenticListed(client, "")
	g.Expect(status).To(gomega.Equal(http.StatusOK))
	g.Expect(names).To(gomega.ConsistOf("managed", "applied"))
	// managed:true.
	names, status = agenticListed(client, "filter=managed:true")
	g.Expect(status).To(gomega.Equal(http.StatusOK))
	g.Expect(names).To(gomega.Equal([]string{"managed"}))
	// managed:false.
	names, status = agenticListed(client, "filter=managed:false")
	g.Expect(status).To(gomega.Equal(http.StatusOK))
	g.Expect(names).To(gomega.Equal([]string{"applied"}))
	// Not a boolean: reported, nothing listed.
	names, _ = agenticListed(client, "filter=managed:maybe")
	g.Expect(names).To(gomega.BeNil())
}
