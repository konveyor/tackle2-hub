package auth

import (
	"context"
	"net/http"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"
)

type BuiltinProvider struct {
	openId *provider.Provider
}

func (p *BuiltinProvider) Handler() http.Handler {
	return p.openId.Handler()
}

// New creates a new OIDC Provider for the Hub
func New() (p *BuiltinProvider, err error) {
	jwksProvider := func(ctx context.Context) (jwks goidc.JSONWebKeySet, err error) {
		// TODO: Generate or load RSA/ECDSA private key properly
		// Example static JWKS (replace with real implementation)
		jwks = goidc.JSONWebKeySet{
			Keys: []goidc.JSONWebKey{
				{
					KeyID:     "kid-1",
					Algorithm: "RS256",
					// Key: yourPrivateKey,   // TODO: add actual key here
				},
			},
		}
		return
	}
	authPolicy := goidc.NewPolicy(
		"main",
		func(r *http.Request, client *goidc.Client, session *goidc.AuthnSession) bool {
			return true // apply to all requests for now
		},
		func(w http.ResponseWriter, r *http.Request, as *goidc.AuthnSession) (status goidc.Status, err error) {
			// TODO: Full authentication + authorization logic goes here:
			// 1. Check if this is local login (username/password)
			// 2. Or external IdP delegation (if idp=xxx parameter exists)
			// 3. Lookup user from DB
			// 4. Load roles → permissions → scopes
			// 5. Set as.Subject and as.Scopes accordingly
			status = goidc.StatusSuccess
			return
		},
	)
	p.openId, err = provider.New(
		goidc.ProfileOpenID,
		Settings.Auth.Token.Key,
		jwksProvider,
		provider.WithGrantTypes(
			goidc.GrantAuthorizationCode,
			goidc.GrantRefreshToken,
		),
		provider.WithPKCERequired(goidc.CodeChallengeMethodSHA256),
		provider.WithPolicies(authPolicy),
		// provider.WithClientManager(gormClientManager)
		// provider.WithAuthnSessionManager(gormSessionManager)
		// provider.WithGrantManager(gormGrantManager)
	)
	if err != nil {
		return
	}
	return
}
