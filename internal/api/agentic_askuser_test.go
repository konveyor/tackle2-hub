package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
)

// bindAgentRun binds body the way AgentRunCreate does: through BindJSON, which
// is strict and rejects fields the target type does not declare.
func bindAgentRun(body string) (r *AgentRun, err error) {
	ctx := &gin.Context{
		Request: &http.Request{
			Body: io.NopCloser(bytes.NewBufferString(body)),
		},
	}
	h := BaseHandler{}
	r = &AgentRun{}
	err = h.BindJSON(ctx, r)
	return
}

// TestAgentRunAskUserBinding pins spec.execution.askUser through the create
// handler's bind.
//
// AgentRun is an alias for the agentic-controller API type and the bind is
// strict, so whether Hub accepts a field is decided entirely by the version of
// that module it is built against. Built against one that predates askUser,
// Hub answers a run that opts in with 400 `json: unknown field "askUser"`.
//
// The assertion reads the field back out of the marshalled spec rather than
// naming it, deliberately: naming it would turn a downgrade into a build
// failure in this file, where reading it back fails with the same error a
// client would be sent.
func TestAgentRunAskUserBinding(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r, err := bindAgentRun(
		`{"metadata":{"name":"r1"},` +
			`"spec":{"agentRef":"a","execution":{"askUser":true,"maxTurns":5}}}`)
	g.Expect(err).To(gomega.BeNil())

	out, err := json.Marshal(r.Spec.Execution)
	g.Expect(err).To(gomega.BeNil())
	// The inlined limits bind alongside it; askUser is not shadowing them.
	g.Expect(string(out)).To(gomega.MatchJSON(`{"askUser":true,"maxTurns":5}`))
}

// TestAgentRunAskUserDefaultsOff guards the default the field exists to
// provide: a run nobody opted into supervising cannot be failed by a question
// nobody saw. A caller that never mentions askUser must not acquire it, and
// the run reads back without the key, exactly as it did before the field
// existed.
func TestAgentRunAskUserDefaultsOff(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r, err := bindAgentRun(
		`{"metadata":{"name":"r1"},` +
			`"spec":{"agentRef":"a","execution":{"maxTurns":5}}}`)
	g.Expect(err).To(gomega.BeNil())

	out, err := json.Marshal(r.Spec.Execution)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(string(out)).To(gomega.MatchJSON(`{"maxTurns":5}`))
}
