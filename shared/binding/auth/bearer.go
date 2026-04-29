package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

var Scopes = []string{
	"openid",
	"profile",
	"email",
	"offline_access",
}

// NewBearer creates a new OIDC bearer token authenticator.
func NewBearer(issuerURL, clientID string) (h *Bearer, err error) {
	rpClient, err := rp.NewRelyingPartyOIDC(
		context.Background(),
		issuerURL,
		clientID,
		"", // public client, no secret
		"", // no redirect URI for device flow
		Scopes,
		rp.WithHTTPClient(
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}),
	)
	if err != nil {
		return
	}
	h = &Bearer{
		rpClient: rpClient,
	}
	return
}

// Bearer provides OIDC authentication with automatic token refresh.
type Bearer struct {
	mutex        sync.RWMutex
	rpClient     rp.RelyingParty
	accessToken  string
	refreshToken string
}

// Use sets the access token.
func (p *Bearer) Use(token string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.accessToken = token
}

// Token returns the access token.
func (p *Bearer) Token() (token string) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	token = p.accessToken
	return
}

// Login performs authentication and refreshes credentials.
func (p *Bearer) Login() (err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.refreshToken == "" {
		err = fmt.Errorf("not authenticated - call DeviceLogin first")
		return
	}

	ctx := context.Background()
	tokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](ctx, p.rpClient, p.refreshToken, "", "")
	if err != nil {
		return
	}

	p.accessToken = tokens.AccessToken
	p.refreshToken = tokens.RefreshToken
	return
}

// Header returns the Authorization header value.
func (p *Bearer) Header() (header string) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	header = "Bearer " + p.accessToken
	return
}

// DeviceAuth initiates device authorization flow.
func (p *Bearer) DeviceAuth(ctx context.Context, scopes []string) (resp *oidc.DeviceAuthorizationResponse, err error) {
	resp, err = rp.DeviceAuthorization(ctx, scopes, p.rpClient, nil)
	return
}

// DeviceAccessToken polls and exchanges device code for tokens.
func (p *Bearer) DeviceAccessToken(
	ctx context.Context,
	deviceCode string,
	interval time.Duration) (resp *oidc.AccessTokenResponse, err error) {
	//
	resp, err = rp.DeviceAccessToken(ctx, deviceCode, interval, p.rpClient)
	if err != nil {
		return
	}
	// Store tokens
	func() {
		p.mutex.Lock()
		p.mutex.Unlock()
		p.accessToken = resp.AccessToken
		p.refreshToken = resp.RefreshToken
	}()

	return
}

// DeviceLogin performs complete device authorization flow with user interaction.
func (p *Bearer) DeviceLogin(ctx context.Context) (err error) {
	device, err := p.DeviceAuth(ctx, Scopes)
	if err != nil {
		return
	}

	fmt.Printf("\nDevice Authorization:\n")
	fmt.Printf("  Visit: %s\n", device.VerificationURI)
	fmt.Printf("  Enter code: %s\n\n", device.UserCode)

	interval := time.Duration(device.Interval) * time.Second
	if interval == 0 {
		interval = 5 * time.Second
	}

	_, err = p.DeviceAccessToken(ctx, device.DeviceCode, interval)
	return
}
