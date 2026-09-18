package api

import (
	"errors"
	"net/http"
	"testing"

	agent "github.com/konveyor/agentic-controller/api/v1alpha1"
	"github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func agentWithLabels(name string, labels map[string]string) *Agent {
	return &agent.Agent{
		ObjectMeta: v1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func TestOperatorManaged(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := &AgenticHandler{}

	// Operator-installed default content.
	err := h.operatorManaged(
		"agent",
		agentWithLabels(
			"migration-plan-agent",
			map[string]string{
				OperatorManagedByLabel: OperatorDefaultsManager,
			}))
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(errors.Is(err, &Conflict{})).To(gomega.BeTrue())
	// The reason has to name the object and say what to do instead.
	g.Expect(err.Error()).To(gomega.ContainSubstring("migration-plan-agent"))
	g.Expect(err.Error()).To(gomega.ContainSubstring("operator"))
	g.Expect(err.Error()).To(gomega.ContainSubstring("Create your own agent"))
}

func TestOperatorManagedPermitsUserContent(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	h := &AgenticHandler{}

	for _, labels := range []map[string]string{
		nil,
		{},
		// ManagedLabel means "needs Hub context", not "operator-owned": it is
		// injected on everything Hub creates, so treating it as ownership
		// would freeze every user-created object.
		{ManagedLabel: "true"},
		// A different manager owns this one.
		{OperatorManagedByLabel: "kustomize"},
	} {
		err := h.operatorManaged("agent", agentWithLabels("mine", labels))
		g.Expect(err).To(gomega.BeNil(), "labels: %v", labels)
	}
}

func TestConflictRespondsWith409(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	err := &Conflict{Reason: "nope"}
	g.Expect(err.Error()).To(gomega.Equal("nope"))
	g.Expect(errors.Is(err, &Conflict{})).To(gomega.BeTrue())
	// Guard the mapping the ErrorHandler relies on.
	g.Expect(errors.Is(err, &Forbidden{})).To(gomega.BeFalse())
	g.Expect(http.StatusConflict).To(gomega.Equal(409))
}
