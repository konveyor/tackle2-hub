package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	frontend "github.com/konveyor/tackle2-hub/internal/frontend/auth"
	"github.com/konveyor/tackle2-hub/shared/api"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
)

const (
	OIDCSubject = "oidc_subject"
)

//
// DagHandler
//

// DagHandler handles device authorization grant HTTP requests.
type DagHandler struct {
	storage *Storage
}

// OIDCAuth creates an OIDC authenticator for the device verification page.
func (h *DagHandler) OIDCAuth() (auth *OIDCAuth) {
	auth = &OIDCAuth{
		pkceState: make(map[string]*PKCEState),
	}
	return
}

// Verify godoc
// @summary Device authorization verification page.
// @description Display page for user to enter device code.
// @tags auth
// @produce html
// @success 200
// @router /auth/device [get]
//
// Verify displays device authorization verification page.
func (h *DagHandler) Verify(ctx *gin.Context) {
	pageReq := frontend.Request{
		Page:             frontend.DeviceVerify,
		DeviceFormAction: AppendIssuer(ctx.Request, api.DeviceRoute),
	}
	h2 := frontend.Handler{}
	_ = h2.Render(ctx.Writer, pageReq)
}

// VerifySubmit godoc
// @summary Submit device authorization.
// @description Process user authorization for device flow.
// @tags auth
// @accept application/x-www-form-urlencoded
// @produce html
// @success 200
// @router /auth/device [post]
// @param userCode formData string true "User code from device"
//
// VerifySubmit processes device authorization submission.
func (h *DagHandler) VerifySubmit(ctx *gin.Context) {
	userCode := ctx.PostForm("userCode")
	if userCode == "" {
		_ = ctx.Error(&BadRequestError{
			Reason: "userCode not provided.",
		})
		return
	}
	userCode = strings.ToUpper(userCode)
	devAuth, found := h.storage.GetDevAuthByUserCode(userCode)
	if !found {
		_ = ctx.Error(&NotFound{
			Resource: "device authorization",
			Id:       userCode,
		})
		return
	}
	if devAuth.Done() || devAuth.Denied() {
		_ = ctx.Error(&BadRequestError{
			Reason: "userCode already used.",
		})
		return
	}
	if time.Now().After(devAuth.Expiration()) {
		_ = ctx.Error(&BadRequestError{
			Reason: "userCode expired.",
		})
		return
	}
	subject := h.currentUser(ctx)
	err := h.storage.UpdateDevAuth(userCode, subject, true, false, time.Now())
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	pageReq := frontend.Request{
		Page: frontend.DeviceSucceeded,
	}
	h2 := frontend.Handler{}
	_ = h2.Render(ctx.Writer, pageReq)
}

// currentUser returns the authenticated user from the gin context.
// Get OIDC subject from context (set by AuthRequired)
func (h *DagHandler) currentUser(ctx *gin.Context) (user string) {
	subject, exists := ctx.Get(OIDCSubject)
	if exists {
		if s, cast := subject.(string); cast {
			user = s
		}
	}
	return
}

//
// OIDCAuth
//

// OIDCAuth provides OIDC authentication for the device verification page.
// Uses server-side state storage to avoid cookie domain issues when hub
// acts as both IdP and RP.
type OIDCAuth struct {
	mutex     sync.Mutex
	cookies   *httphelper.CookieHandler
	pkceState map[string]*PKCEState
	initOnce  sync.Once
}

// Login initiates the OIDC login flow with PKCE.
func (h *OIDCAuth) Login(ctx *gin.Context) {
	h.ensureCookieHandler()

	state := uuid.New().String()

	b := make([]byte, 32)
	_, _ = rand.Read(b)
	codeVerifier := base64.RawURLEncoding.EncodeToString(b)
	digest := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(digest[:])

	h.storeState(state, codeVerifier)

	issuer := Issuer(ctx.Request)
	authURL, _ := url.Parse(issuer)
	authURL.Path, _ = url.JoinPath(authURL.Path, "/authorize")
	redirectURI := AppendIssuer(ctx.Request, api.DeviceCbRoute)

	query := authURL.Query()
	query.Set("client_id", DevVerifierClientId)
	query.Set("response_type", "code")
	query.Set("redirect_uri", redirectURI)
	query.Set("scope", "openid")
	query.Set("state", state)
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", "S256")
	authURL.RawQuery = query.Encode()

	http.Redirect(ctx.Writer, ctx.Request, authURL.String(), http.StatusFound)
}

// Callback handles the OIDC callback and exchanges code for tokens.
func (h *OIDCAuth) Callback(ctx *gin.Context) {
	state := ctx.Query("state")
	code := ctx.Query("code")

	var pkceState *PKCEState
	var found bool
	func() {
		h.mutex.Lock()
		defer h.mutex.Unlock()
		pkceState, found = h.pkceState[state]
		if found {
			delete(h.pkceState, state)
		}
	}()
	if !found {
		_ = ctx.Error(&BadRequestError{
			Reason: "Invalid state parameter",
		})
		return
	}

	issuer := Issuer(ctx.Request)
	tokenURL, _ := url.Parse(issuer)
	tokenURL.Path, _ = url.JoinPath(tokenURL.Path, "/token")

	formData := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {DevVerifierClientId},
		"redirect_uri":  {AppendIssuer(ctx.Request, api.DeviceCbRoute)},
		"code_verifier": {pkceState.verifier},
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Post(
		tokenURL.String(),
		"application/x-www-form-urlencoded",
		bytes.NewBufferString(formData.Encode()))
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		_ = ctx.Error(&BadRequestError{Reason: "Token exchange failed"})
		return
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenResp.IDToken, jwt.MapClaims{})
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	subject := claims["sub"].(string)

	err = h.cookies.SetCookie(ctx.Writer, OIDCSubject, subject)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	redirect := AppendIssuer(ctx.Request, api.DeviceRoute)
	http.Redirect(ctx.Writer, ctx.Request, redirect, http.StatusFound)
}

// AuthRequired checks for valid OIDC session.
func (h *OIDCAuth) AuthRequired(ctx *gin.Context) {
	h.ensureCookieHandler()

	subject, err := h.cookies.CheckCookie(ctx.Request, OIDCSubject)
	if err != nil || subject == "" {
		// No session
		ctx.Redirect(
			http.StatusFound,
			AppendIssuer(ctx.Request, api.DeviceLoginRoute))
		ctx.Abort()
		return
	}

	// Store subject in context for DagHandler
	ctx.Set(OIDCSubject, subject)
	ctx.Next()
}

// ensureCookieHandler initializes the cookie handler if not already done.
func (h *OIDCAuth) ensureCookieHandler() {
	h.initOnce.Do(func() {
		secret := Settings.Auth.APIKey.Secret
		hashKey := h.hashKey256([]byte(secret + "-hash"))
		encryptKey := h.hashKey256([]byte(secret + "-encrypt"))

		h.cookies = httphelper.NewCookieHandler(
			hashKey,
			encryptKey,
			httphelper.WithUnsecure(),
			httphelper.WithSameSite(http.SameSiteLaxMode),
		)
	})
}

// storeState stores PKCE state and cleans up expired states.
func (h *OIDCAuth) storeState(state, verifier string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.pkceState[state] = &PKCEState{
		verifier: verifier,
		created:  time.Now(),
	}

	// Clean up old states (>10 minutes)
	now := time.Now()
	for s, ps := range h.pkceState {
		if now.Sub(ps.created) > 10*time.Minute {
			delete(h.pkceState, s)
		}
	}
}

// hashKey256 derives a 32-byte key using SHA256.
func (h *OIDCAuth) hashKey256(data []byte) (key []byte) {
	hash := sha256.Sum256(data)
	key = hash[:]
	return
}

// PKCEState holds PKCE verifier and metadata.
type PKCEState struct {
	verifier string
	created  time.Time
}
