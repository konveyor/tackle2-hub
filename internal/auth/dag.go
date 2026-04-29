package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
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
	formAction := Settings.Auth.AppendIssuer("/device")
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Device Authorization</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            padding: 40px;
            max-width: 400px;
            width: 100%;
        }
        h1 {
            color: #333;
            font-size: 20px;
            margin-bottom: 8px;
            text-align: center;
        }
        .subtitle {
            color: #666;
            font-size: 13px;
            margin-bottom: 30px;
            text-align: center;
        }
        label {
            display: block;
            color: #333;
            font-size: 13px;
            font-weight: 500;
            margin-bottom: 8px;
        }
        input[type="text"] {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            font-family: monospace;
            letter-spacing: 2px;
            text-transform: uppercase;
            transition: border-color 0.2s;
        }
        input[type="text"]:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            width: 100%;
            padding: 12px;
            margin-top: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 15px;
            font-weight: 500;
            cursor: pointer;
            transition: transform 0.1s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Device Authorization</h1>
        <p class="subtitle">Enter the code displayed on your device</p>
        <form method="POST" action="` + formAction + `">
            <label for="userCode">User Code</label>
            <input type="text" id="userCode" name="userCode" placeholder="XXXX-XXXX" required autofocus>
            <button type="submit">Authorize Device</button>
        </form>
    </div>
</body>
</html>
`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
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

	// Normalize to uppercase for case-insensitive comparison
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

	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Authorization Complete</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            padding: 40px;
            max-width: 400px;
            width: 100%;
            text-align: center;
        }
        .checkmark {
            width: 64px;
            height: 64px;
            border-radius: 50%;
            background: #4caf50;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 32px;
            color: white;
        }
        h1 {
            color: #333;
            font-size: 20px;
            margin-bottom: 16px;
        }
        p {
            color: #666;
            font-size: 13px;
            line-height: 1.6;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="checkmark">✓</div>
        <h1>Authorization Complete</h1>
        <p>You have successfully authorized the device.<br>You may close this window.</p>
    </div>
</body>
</html>
`
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
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
	mutex     sync.RWMutex
	rpClient  rp.RelyingParty
	cookies   *httphelper.CookieHandler
	pkceState map[string]*PKCEState
	initOnce  sync.Once
}

// Login initiates the OIDC login flow with PKCE.
func (h *OIDCAuth) Login(ctx *gin.Context) {
	err := h.ensureRpClient()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	state := uuid.New().String()

	// build verifier and challenge
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	codeVerifier := base64.RawURLEncoding.EncodeToString(b)
	digest := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(digest[:])

	// Store state and verifier
	h.storeState(state, codeVerifier)

	// Build authorize URL with PKCE
	authURL := rp.AuthURL(state, h.rpClient, rp.WithCodeChallenge(codeChallenge))

	http.Redirect(ctx.Writer, ctx.Request, authURL, http.StatusFound)
}

// Callback handles the OIDC callback and exchanges code for tokens.
func (h *OIDCAuth) Callback(ctx *gin.Context) {
	err := h.ensureRpClient()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	state := ctx.Query("state")
	code := ctx.Query("code")

	// Retrieve and validate state
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

	// Exchange code for tokens with PKCE verifier
	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](
		ctx.Request.Context(),
		code,
		h.rpClient,
		rp.WithCodeVerifier(pkceState.verifier),
	)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	subject := tokens.IDTokenClaims.Subject

	// Store subject in cookie.
	err = h.cookies.SetCookie(ctx.Writer, OIDCSubject, subject)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// Redirect to device authorization page
	http.Redirect(
		ctx.Writer,
		ctx.Request,
		Settings.Auth.AppendIssuer("/device"),
		http.StatusFound)
}

// AuthRequired checks for valid OIDC session.
func (h *OIDCAuth) AuthRequired(ctx *gin.Context) {
	err := h.ensureRpClient()
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	// Check for session cookie
	subject, err := h.cookies.CheckCookie(ctx.Request, OIDCSubject)
	if err != nil || subject == "" {
		// No session
		ctx.Redirect(
			http.StatusFound,
			Settings.Auth.AppendIssuer("/device/login"))
		ctx.Abort()
		return
	}

	// Store subject in context for DagHandler
	ctx.Set(OIDCSubject, subject)
	ctx.Next()
}

// ensureRpClient initializes the RP client if not already done.
func (h *OIDCAuth) ensureRpClient() (err error) {
	h.initOnce.Do(func() {
		issuer := Settings.Auth.IssuerURL

		// Derive keys
		secret := Settings.Auth.APIKey.Secret
		hashKey := h.hashKey256([]byte(secret + "-hash"))
		encryptKey := h.hashKey256([]byte(secret + "-encrypt"))

		// Create cookie handler for session management
		h.cookies = httphelper.NewCookieHandler(
			hashKey,
			encryptKey,
			httphelper.WithUnsecure(),
			httphelper.WithSameSite(http.SameSiteLaxMode),
		)

		// Create OIDC RP client (no secret - internal client)
		h.rpClient, err = rp.NewRelyingPartyOIDC(
			context.Background(),
			issuer,
			DevVerifierClientId,
			"",
			Settings.Auth.AppendIssuer("/device/callback"),
			[]string{"openid"},
			rp.WithHTTPClient(
				&http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				}),
		)
	})
	return
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
