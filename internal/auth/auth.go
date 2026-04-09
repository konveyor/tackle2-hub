package auth

import (
	"errors"
	"net/http"
	"sort"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewAuthManager returns an authn manager.
func NewAuthManager(db *gorm.DB) (m *AuthManager) {
	m = &AuthManager{db: db}
	return
}

// AuthManager applies authN and AuthZ.
type AuthManager struct {
	db *gorm.DB
}

// Login provides the access token authentication.
func (r *AuthManager) Login(
	writer http.ResponseWriter,
	request *http.Request,
	session *goidc.AuthnSession) (status goidc.Status, err error) {
	//
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	var userid, password string
	if session.Subject == "" {
		userid = request.PostFormValue("userid")
		password = request.PostFormValue("password")
	}
	if userid == "" || password == "" {
		err = r.renderPage(writer, request, session)
		status = goidc.StatusInProgress
		return
	}
	user := &model.User{}
	db := r.db.Preload(clause.Associations)
	db = db.Preload("Roles.Permissions")
	err = db.First(user, "userid", userid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.renderPage(writer, request, session)
			status = goidc.StatusInProgress
			err = nil
		}
		return
	}
	if !secret.MatchPassword(password, user.Password) {
		err = r.renderPage(writer, request, session)
		status = goidc.StatusInProgress
		return
	}
	r.appendScopes(session, user)
	r.grantResources(session)
	session.Subject = user.Subject
	status = goidc.StatusSuccess
	return
}

// appendScopes grants requested scopes plus additional scopes from user roles.
func (r *AuthManager) appendScopes(session *goidc.AuthnSession, user *model.User) {
	unique := make(map[string]byte)
	// Start with requested scopes
	for _, scope := range strings.Fields(session.Scopes) {
		unique[scope] = 0
	}
	// Add scopes from user roles
	for _, role := range user.Roles {
		unique["role="+role.Name] = 0
		for _, permission := range role.Permissions {
			unique[permission.Scope] = 0
		}
	}
	scopes := make([]string, 0, len(unique))
	for scope := range unique {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	session.GrantedScopes = strings.Join(scopes, " ")
	return
}

// grantResources grants resources for the token audience.
// If the client requested specific resources, grant those.
// Otherwise, grant the issuer URL as the default audience.
func (r *AuthManager) grantResources(session *goidc.AuthnSession) {
	resources := session.Resources
	if len(resources) == 0 {
		issuer := Settings.IssuerURL
		if issuer == "" {
			issuer = Settings.Addon.Hub.URL + api.OIDCRoutes
		}
		resources = []string{issuer}
	}
	session.GrantResources(resources)
}

// renderPage renders the login page.
func (r *AuthManager) renderPage(writer http.ResponseWriter, _ *http.Request, session *goidc.AuthnSession) (err error) {
	defer func() {
		if err != nil {
			Log.Error(err, "")
		}
	}()
	issuer := Settings.IssuerURL
	if issuer == "" {
		issuer = Settings.Addon.Hub.URL + api.OIDCRoutes
	}
	// Simple login form HTML - POST to callback URL with session CallbackID
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Tackle Hub - Login</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 100px auto; padding: 20px; }
        h1 { color: #333; }
        form { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        input { width: 100%; padding: 8px; margin: 10px 0; box-sizing: border-box; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; width: 100%; }
        button:hover { background: #0056b3; }
    </style>
</head>
<body>
    <h1>Tackle Login</h1>
    <form action="` + issuer + `/authorize/` + session.CallbackID + `" method="post">
        <div>
            <label>Userid:</label>
            <input type="text" name="userid" required autofocus />
        </div>
        <div>
            <label>Password:</label>
            <input type="password" name="password" required />
        </div>
        <button type="submit">Login</button>
    </form>
</body>
</html>`
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = writer.Write([]byte(html))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
