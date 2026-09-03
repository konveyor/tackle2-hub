package agentic

const (
	// TokenSecretLabel identifies Secrets containing short-lived Hub
	// credentials minted for an AgentRun or AgentWorkflowRun.
	TokenSecretLabel = "konveyor.io/agentic-token"

	// TokenSecretFinalizer keeps the token ID available until the Hub janitor
	// has revoked the corresponding database token.
	TokenSecretFinalizer = "tackle.konveyor.io/agentic-token"

	TokenIDKey = "HUB_TOKEN_ID"
	TokenKey   = "HUB_TOKEN"
)
