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

	secret := "test-secret"

	// Create API client
	client := NewClient()

	// Create test OAuth client with client_credentials grant
	testClient := &api.IdpClient{
		ClientId:        "test-client-credentials",
		Secret:          secret,
		ApplicationType: "web",
		Grants:          []string{"client_credentials"},
		Scopes:          []string{"openid", "profile", "email", "applications:get"},
	}
	err := client.IdpClient.Create(testClient)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.IdpClient.Delete(testClient.ID)
	})

	// Request token using client credentials
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", testClient.ClientId)
	form.Set("client_secret", secret)
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
		Scope        string `json:"scope"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	g.Expect(err).To(BeNil())
	g.Expect(tokenResp.AccessToken).NotTo(BeEmpty())
	g.Expect(tokenResp.TokenType).To(Equal("Bearer"))
	g.Expect(tokenResp.ExpiresIn).To(BeNumerically(">", 0))
	t.Logf("Token scopes: %s", tokenResp.Scope)

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

	secret := "test-secret"

	// Create API client
	client := NewClient()

	// Create test OAuth client with authorization_code grant
	testClient := &api.IdpClient{
		ClientId:        "test-authorization-code",
		Secret:          secret,
		ApplicationType: "web",
		Grants:          []string{"authorization_code", "refresh_token"},
		RedirectURIs:    []string{"http://test-redirect"},
		Scopes:          []string{"openid", "profile", "email"},
	}
	err := client.IdpClient.Create(testClient)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.IdpClient.Delete(testClient.ID)
	})

	// Use the configured issuer URL (not from discovery, which may have different hostname)
	issuer := Settings.Addon.Hub.URL + api.OIDCRoutes

	// Create test user with admin role
	user := api.User{
		Login:    "oidc-test-user",
		Email:    "oidc-test@example.com",
		Password: "oidc-test-password",
		Roles: []api.Ref{
			{ID: 1}, // admin role
		},
	}
	err = client.User.Create(&user)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.User.Delete(user.ID)
	})

	// Use the original plaintext password for login (not the encrypted one from API)
	username := user.Login
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
	var loginURL string
	redirectURI := "http://test-redirect"
	authURL := issuer + "/authorize?" +
		"client_id=" + testClient.ClientId +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&response_type=code" +
		"&scope=openid+profile+email" +
		"&code_challenge=" + challenge +
		"&code_challenge_method=S256"

	resp, err := httpClient.Get(authURL)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Should get redirect to login page
	g.Expect(resp.StatusCode).To(Equal(http.StatusFound))
	loginURL = resp.Header.Get("Location")
	g.Expect(loginURL).To(ContainSubstring("/login?authRequestId="))

	// Follow redirect to get login page
	resp, err = httpClient.Get(loginURL)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	g.Expect(html).To(ContainSubstring("Tackle Login"))

	// Extract auth request ID from URL
	parsedURL, err := url.Parse(loginURL)
	g.Expect(err).To(BeNil())
	authReqID := parsedURL.Query().Get("authRequestId")
	g.Expect(authReqID).NotTo(BeEmpty())

	// Step 2: Submit login form (POST /login?authRequestId=...)
	loginURL = issuer + "/login?authRequestId=" + authReqID
	loginForm := url.Values{}
	loginForm.Set("login", username)
	loginForm.Set("password", password)

	resp, err = httpClient.PostForm(loginURL, loginForm)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Should redirect to provider callback (302 or 303)
	g.Expect(resp.StatusCode).To(BeNumerically(">=", 300))
	g.Expect(resp.StatusCode).To(BeNumerically("<", 400))
	callbackLocation := resp.Header.Get("Location")
	g.Expect(callbackLocation).To(ContainSubstring("/authorize/callback"))

	// Follow redirect to callback to get authorization code
	resp, err = httpClient.Get(callbackLocation)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Callback should redirect to client redirect_uri with code
	g.Expect(resp.StatusCode).To(BeNumerically(">=", 300))
	g.Expect(resp.StatusCode).To(BeNumerically("<", 400))
	location := resp.Header.Get("Location")
	g.Expect(location).To(ContainSubstring("?code="))

	// Extract authorization code
	parsedURL, err = url.Parse(location)
	g.Expect(err).To(BeNil())
	code := parsedURL.Query().Get("code")
	g.Expect(code).NotTo(BeEmpty())

	// Step 3: Exchange code for tokens (POST /oauth/token)
	tokenForm := url.Values{}
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", code)
	tokenForm.Set("redirect_uri", redirectURI)
	tokenForm.Set("client_id", testClient.ClientId)
	tokenForm.Set("client_secret", secret)
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

// TestAuthorizationCodeFlowWithScopes tests that user roles inject scopes into tokens.
func TestAuthorizationCodeFlowWithScopes(t *testing.T) {
	g := NewGomegaWithT(t)

	secret := "test-secret"

	// Create API client
	client := NewClient()

	// Create test OAuth client with authorization_code grant
	testClient := &api.IdpClient{
		ClientId:        "test-authorization-scopes",
		Secret:          secret,
		ApplicationType: "web",
		Grants:          []string{"authorization_code", "refresh_token"},
		RedirectURIs:    []string{"http://test-redirect"},
		Scopes:          []string{"openid", "profile", "email"},
	}
	err := client.IdpClient.Create(testClient)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.IdpClient.Delete(testClient.ID)
	})

	permissions, err := client.Permission.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(permissions)).To(BeNumerically(">", 0))

	// Find applications:get permission
	var appsGetPerm *api.Permission
	for i := range permissions {
		if permissions[i].Scope == "applications:get" {
			appsGetPerm = &permissions[i]
			break
		}
	}
	g.Expect(appsGetPerm).NotTo(BeNil(), "applications:get permission not found")

	// Create role with multiple permissions including applications:get
	role := api.Role{
		Name: "Test Role",
		Permissions: []api.Ref{
			{ID: permissions[0].ID, Name: permissions[0].Name}, // First permission (for scope test)
			{ID: appsGetPerm.ID, Name: appsGetPerm.Name},       // applications:get
		},
	}
	err = client.Role.Create(&role)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.Role.Delete(role.ID)
	})

	// Create test user with role
	user := api.User{
		Login:    "oidc-role-test-user",
		Email:    "oidc-role-test@example.com",
		Password: "oidc-role-test-password",
		Roles: []api.Ref{
			{ID: role.ID, Name: role.Name},
		},
	}
	err = client.User.Create(&user)
	g.Expect(err).To(BeNil())
	t.Cleanup(func() {
		_ = client.User.Delete(user.ID)
	})

	// Use the configured issuer URL (not from discovery, which may have different hostname)
	issuer := Settings.Addon.Hub.URL + api.OIDCRoutes

	// Use plaintext password
	username := user.Login
	password := "oidc-role-test-password"

	// Create HTTP client with cookie jar
	jar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Generate PKCE challenge
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// Step 1: Request authorization
	var loginURL string
	redirectURI := "http://test-redirect"
	authURL := issuer + "/authorize?" +
		"client_id=" + testClient.ClientId +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&response_type=code" +
		"&scope=openid+profile+email" +
		"&code_challenge=" + challenge +
		"&code_challenge_method=S256"

	resp, err := httpClient.Get(authURL)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Should get redirect to login page
	g.Expect(resp.StatusCode).To(Equal(http.StatusFound))
	loginURL = resp.Header.Get("Location")
	g.Expect(loginURL).To(ContainSubstring("/login?authRequestId="))

	// Follow redirect to get login page
	resp, err = httpClient.Get(loginURL)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	g.Expect(html).To(ContainSubstring("Tackle Login"))

	// Extract auth request ID from URL
	parsedURL, err := url.Parse(loginURL)
	g.Expect(err).To(BeNil())
	authReqID := parsedURL.Query().Get("authRequestId")
	g.Expect(authReqID).NotTo(BeEmpty())

	// Step 2: Submit login form
	loginURL = issuer + "/login?authRequestId=" + authReqID
	loginForm := url.Values{}
	loginForm.Set("login", username)
	loginForm.Set("password", password)

	resp, err = httpClient.PostForm(loginURL, loginForm)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Should redirect to provider callback
	g.Expect(resp.StatusCode).To(BeNumerically(">=", 300))
	g.Expect(resp.StatusCode).To(BeNumerically("<", 400))
	callbackLocation := resp.Header.Get("Location")
	g.Expect(callbackLocation).To(ContainSubstring("/authorize/callback"))

	// Follow redirect to callback to get authorization code
	resp, err = httpClient.Get(callbackLocation)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Callback should redirect to client redirect_uri with code
	g.Expect(resp.StatusCode).To(BeNumerically(">=", 300))
	g.Expect(resp.StatusCode).To(BeNumerically("<", 400))
	location := resp.Header.Get("Location")

	parsedURL, err = url.Parse(location)
	g.Expect(err).To(BeNil())
	code := parsedURL.Query().Get("code")
	g.Expect(code).NotTo(BeEmpty())

	// Step 3: Exchange code for tokens
	tokenForm := url.Values{}
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", code)
	tokenForm.Set("redirect_uri", redirectURI)
	tokenForm.Set("client_id", testClient.ClientId)
	tokenForm.Set("client_secret", secret)
	tokenForm.Set("code_verifier", verifier)

	resp, err = http.PostForm(issuer+"/token", tokenForm)
	g.Expect(err).To(BeNil())
	defer resp.Body.Close()

	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	g.Expect(err).To(BeNil())
	g.Expect(tokenResp.AccessToken).NotTo(BeEmpty())

	// Decode JWT to verify scopes (base64 decode the payload, no signature verification needed)
	parts := strings.Split(tokenResp.AccessToken, ".")
	g.Expect(parts).To(HaveLen(3))

	// Decode payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	g.Expect(err).To(BeNil())

	var claims map[string]interface{}
	err = json.Unmarshal(payload, &claims)
	g.Expect(err).To(BeNil())

	// Verify scopes are in the token
	scopeRaw, ok := claims["scope"].(string)
	g.Expect(ok).To(BeTrue())
	g.Expect(scopeRaw).To(ContainSubstring("addons:delete"))

	// Test using the access token
	req, _ := http.NewRequest("GET", Settings.Addon.Hub.URL+"/applications", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	apiResp, err := http.DefaultClient.Do(req)
	g.Expect(err).To(BeNil())
	defer apiResp.Body.Close()
	g.Expect(apiResp.StatusCode).To(Equal(http.StatusOK))
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
