package api

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
)

// TestAgentRunAskUserBinding pins spec.execution.askUser through the same
// unmarshal the create handler performs.
//
// AgentRun is an alias for the agentic-controller API type, so the request
// body is bound straight into it. A field the pinned module does not know
// about is not an error -- it is discarded in silence, and the run is created
// without it. That is how this one failed before: the console sent askUser,
// Hub answered 201, and the agent ran without the ask_user tool with nothing
// anywhere to say why. A downgrade of that module past this field would
// reintroduce exactly that, so assert the round-trip rather than trusting the
// pin.
func TestAgentRunAskUserBinding(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &AgentRun{}
	body := `{"metadata":{"name":"r1"},` +
		`"spec":{"agentRef":"a","execution":{"askUser":true,"maxTurns":5}}}`
	g.Expect(json.Unmarshal([]byte(body), r)).To(gomega.Succeed())
	g.Expect(r.Spec.Execution).ToNot(gomega.BeNil())
	g.Expect(r.Spec.Execution.AskUser).To(
		gomega.BeTrue(),
		"askUser was dropped binding the request: the agentic-controller API"+
			" module predates spec.execution.askUser")
	// The inlined limits still bind; askUser is not shadowing them.
	g.Expect(r.Spec.Execution.MaxTurns).To(gomega.HaveValue(gomega.Equal(5)))
}

// TestAgentRunAskUserDefaultsOff guards the default the field exists to
// provide: a run nobody opted into supervising cannot be failed by an
// unanswered question. Callers that never mention askUser must get false.
func TestAgentRunAskUserDefaultsOff(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := &AgentRun{}
	body := `{"metadata":{"name":"r1"},` +
		`"spec":{"agentRef":"a","execution":{"maxTurns":5}}}`
	g.Expect(json.Unmarshal([]byte(body), r)).To(gomega.Succeed())
	g.Expect(r.Spec.Execution).ToNot(gomega.BeNil())
	g.Expect(r.Spec.Execution.AskUser).To(gomega.BeFalse())

	// omitempty: an opted-out run carries no askUser key on the way back out,
	// so it reads the same to clients as it did before the field existed.
	out, err := json.Marshal(r.Spec.Execution)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(string(out)).ToNot(gomega.ContainSubstring("askUser"))
}
