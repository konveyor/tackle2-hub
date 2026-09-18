package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/onsi/gomega"
)

// TestAgentRunBindsAskUser guards the reason this handler's AgentRun type is
// pinned to an agentic-controller API that carries spec.execution.askUser
// (agentic-controller#242): the agent's ask_user tool is opt-in, and the
// console asks for it per run. Binding here is strict
// (BaseHandler.BindJSON sets DisallowUnknownFields), so an API without the
// field does not quietly ignore it — the whole create fails with 400 and no
// run is made. Decoding with the same settings keeps that failure a test
// failure instead of a rejected request.
func TestAgentRunBindsAskUser(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	body := `{
	  "metadata": {"generateName": "ui-"},
	  "spec": {
	    "agentRef": "coolstore-quarkus-migrator",
	    "execution": {"mode": "auto", "askUser": true}
	  }
	}`

	run := &AgentRun{}
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.DisallowUnknownFields()
	g.Expect(decoder.Decode(run)).To(gomega.Succeed())

	g.Expect(run.Spec.Execution).ToNot(gomega.BeNil())
	g.Expect(run.Spec.Execution.AskUser).To(gomega.BeTrue())
	g.Expect(string(run.Spec.Execution.Mode)).To(gomega.Equal("auto"))

	// Omitted is off: a run nobody opted into supervising cannot be failed
	// by a question nobody saw.
	quiet := &AgentRun{}
	decoder = json.NewDecoder(strings.NewReader(
		`{"spec": {"agentRef": "a", "execution": {"mode": "auto"}}}`))
	decoder.DisallowUnknownFields()
	g.Expect(decoder.Decode(quiet)).To(gomega.Succeed())
	g.Expect(quiet.Spec.Execution.AskUser).To(gomega.BeFalse())
}
