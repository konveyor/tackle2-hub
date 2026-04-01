package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	. "github.com/onsi/gomega"
)

// TestOIDCDiscovery tests the OpenID Connect discovery endpoint.
func TestOIDCDiscovery(t *testing.T) {
	g := NewGomegaWithT(t)

	// Request discovery document
	resp, err := http.Get(Settings.Addon.Hub.URL + api.OIDCRoutes + "/.well-known/openid-configuration")
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var discovery map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&discovery)
	g.Expect(err).To(BeNil())

	// Verify required OIDC endpoints
	g.Expect(discovery["issuer"]).NotTo(BeEmpty())
	g.Expect(discovery["authorization_endpoint"]).NotTo(BeEmpty())
	g.Expect(discovery["token_endpoint"]).NotTo(BeEmpty())
	g.Expect(discovery["jwks_uri"]).NotTo(BeEmpty())
	g.Expect(discovery["userinfo_endpoint"]).NotTo(BeEmpty())

	// Verify supported grant types
	grantTypes := discovery["grant_types_supported"].([]interface{})
	g.Expect(grantTypes).To(ContainElement("client_credentials"))
	g.Expect(grantTypes).To(ContainElement("authorization_code"))
	g.Expect(grantTypes).To(ContainElement("refresh_token"))

	// Verify supported scopes
	scopes := discovery["scopes_supported"].([]interface{})
	g.Expect(scopes).To(ContainElement("openid"))
	g.Expect(scopes).To(ContainElement("profile"))
	g.Expect(scopes).To(ContainElement("email"))

	// Verify PKCE is required
	pkce := discovery["code_challenge_methods_supported"].([]interface{})
	g.Expect(pkce).To(ContainElement("S256"))
}

// TestClientCredentialsFlow tests the machine-to-machine OAuth2 flow.
func TestClientCredentialsFlow(t *testing.T) {
	g := NewGomegaWithT(t)

	// Request token using client credentials
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", Settings.Auth.Client.ID)
	form.Set("client_secret", Settings.Auth.Client.Secret)
	form.Set("scope", "openid")

	resp, err := http.PostForm(Settings.Addon.Hub.URL+api.OIDCRoutes+"/token", form)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Verify response
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	g.Expect(err).To(BeNil())
	g.Expect(tokenResp.AccessToken).NotTo(BeEmpty())
	g.Expect(tokenResp.TokenType).To(Equal("Bearer"))
	g.Expect(tokenResp.ExpiresIn).To(BeNumerically(">", 0))

	// Test using the access token to call an API endpoint
	req, _ := http.NewRequest("GET", Settings.Addon.Hub.URL+"/applications", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	apiResp, err := http.DefaultClient.Do(req)
	g.Expect(err).To(BeNil())
	defer apiResp.Body.Close()
	g.Expect(apiResp.StatusCode).To(Equal(http.StatusOK))
}

// TestAuthorizationCodeFlow tests the user-based OAuth2 flow with PKCE.
func TestAuthorizationCodeFlow(t *testing.T) {
	g := NewGomegaWithT(t)

	// Get the issuer URL from discovery document
	resp, err := http.Get(Settings.Addon.Hub.URL + api.OIDCRoutes + "/.well-known/openid-configuration")
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	var discovery map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&discovery)
	g.Expect(err).To(BeNil())
	issuer := discovery["issuer"].(string)

	// Create test user
	user := api.User{
		Name:     "oidc-test-user",
		Email:    "oidc-test@example.com",
		Password: "oidc-test-password",
	}
	err = client.User.Create(&user)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.User.Delete(user.ID)
	})

	// Use the original plaintext password for login (not the encrypted one from API)
	username := user.Name
	password := "oidc-test-password"

	// Create HTTP client with cookie jar (to maintain session)
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically
			return http.ErrUseLastResponse
		},
	}

	// Generate PKCE challenge
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// Step 1: Request authorization (GET /authorize)
	// Use the issuer URL for redirect_uri to match what's configured in the client
	redirectURI := issuer + "/callback"
	authURL := issuer + "/authorize?" +
		"client_id=" + Settings.Auth.Client.ID +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&response_type=code" +
		"&scope=openid+profile+email" +
		"&code_challenge=" + challenge +
		"&code_challenge_method=S256"

	resp, err = httpClient.Get(authURL)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Should get login page
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	g.Expect(html).To(ContainSubstring("Tackle Hub Login"))

	// Extract CallbackID from form action
	callbackID := extractCallbackID(html)
	g.Expect(callbackID).NotTo(BeEmpty())

	// Step 2: Submit login form (POST /authorize/{callbackID})
	loginURL := issuer + "/authorize/" + callbackID
	loginForm := url.Values{}
	loginForm.Set("userid", username)
	loginForm.Set("password", password)

	resp, err = httpClient.PostForm(loginURL, loginForm)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Should redirect to callback with authorization code (302 or 303)
	g.Expect(resp.StatusCode).To(BeNumerically(">=", 300))
	g.Expect(resp.StatusCode).To(BeNumerically("<", 400))
	location := resp.Header.Get("Location")
	g.Expect(location).To(ContainSubstring("/callback?code="))

	// Extract authorization code
	parsedURL, err := url.Parse(location)
	g.Expect(err).To(BeNil())
	code := parsedURL.Query().Get("code")
	g.Expect(code).NotTo(BeEmpty())

	// Step 3: Exchange code for tokens (POST /token)
	tokenForm := url.Values{}
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", code)
	tokenForm.Set("redirect_uri", redirectURI)
	tokenForm.Set("client_id", Settings.Auth.Client.ID)
	tokenForm.Set("client_secret", Settings.Auth.Client.Secret)
	tokenForm.Set("code_verifier", verifier)

	resp, err = http.PostForm(issuer+"/token", tokenForm)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Verify token response
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	g.Expect(err).To(BeNil())
	g.Expect(tokenResp.AccessToken).NotTo(BeEmpty())
	g.Expect(tokenResp.TokenType).To(Equal("Bearer"))
	g.Expect(tokenResp.ExpiresIn).To(BeNumerically(">", 0))

	// Test using the access token
	req, _ := http.NewRequest("GET", Settings.Addon.Hub.URL+"/applications", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	apiResp, err := http.DefaultClient.Do(req)
	g.Expect(err).To(BeNil())
	defer apiResp.Body.Close()
	g.Expect(apiResp.StatusCode).To(Equal(http.StatusOK))
}

// TestRefreshTokenFlow tests token refresh.
func TestRefreshTokenFlow(t *testing.T) {
	g := NewGomegaWithT(t)

	// First get a refresh token via client_credentials
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", Settings.Auth.Client.ID)
	form.Set("client_secret", Settings.Auth.Client.Secret)
	form.Set("scope", "openid")

	resp, err := http.PostForm(Settings.Addon.Hub.URL+api.OIDCRoutes+"/token", form)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	var initialTokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&initialTokenResp)
	g.Expect(err).To(BeNil())
	g.Expect(initialTokenResp.AccessToken).NotTo(BeEmpty())

	// Skip refresh token test if not provided
	if initialTokenResp.RefreshToken == "" {
		t.Skip("Refresh token not provided in client_credentials response")
		return
	}

	// Use refresh token to get new access token
	refreshForm := url.Values{}
	refreshForm.Set("grant_type", "refresh_token")
	refreshForm.Set("refresh_token", initialTokenResp.RefreshToken)
	refreshForm.Set("client_id", Settings.Auth.Client.ID)
	refreshForm.Set("client_secret", Settings.Auth.Client.Secret)

	resp, err = http.PostForm(Settings.Addon.Hub.URL+api.OIDCRoutes+"/token", refreshForm)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var newTokenResp struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&newTokenResp)
	g.Expect(err).To(BeNil())
	g.Expect(newTokenResp.AccessToken).NotTo(BeEmpty())
}

// extractCallbackID extracts the callback ID from the login form HTML.
func extractCallbackID(html string) string {
	// Find form action attribute
	start := strings.Index(html, `action="`)
	if start == -1 {
		return ""
	}
	start += len(`action="`)
	end := strings.Index(html[start:], `"`)
	if end == -1 {
		return ""
	}

	action := html[start : start+end]
	// Extract callback ID from URL like: /oidc/authorize/abc123
	parts := strings.Split(action, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
