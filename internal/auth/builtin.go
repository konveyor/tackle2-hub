package auth

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	cache2 "github.com/konveyor/tackle2-hub/internal/auth/cache"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/gorm"
)

// NewBuiltin returns a configured provider.
func NewBuiltin(db *gorm.DB) (builtin *Builtin, err error) {
	cache := cache2.New(db)
	builtin = &Builtin{
		cache: cache,
		db:    db,
	}
	keyManager := NewKeyManager(db)
	builtin.keySet, err = keyManager.KeySet()
	if err != nil {
		return
	}
	builtin.storage = &Storage{
		db:                  db,
		keySet:              builtin.keySet,
		authReqById:         make(map[string]*AuthRequest),
		authReqByCode:       make(map[string]string),
		devAuthReqByDevCode: make(map[string]*DeviceAuthRequest),
		devAuthByUserCode:   make(map[string]string),
		cache:               cache,
	}
	basePath := api.OIDCRoutes
	config := &op.Config{
		CodeMethodS256:          true,
		AuthMethodPost:          true,
		AuthMethodPrivateKeyJWT: false,
		GrantTypeRefreshToken:   true,
		RequestObjectSupported:  false,
		DeviceAuthorization: op.DeviceAuthorizationConfig{
			Lifetime:     15 * time.Minute,
			PollInterval: 5 * time.Second,
			UserFormPath: basePath + api.DeviceRoute,
			UserCode: op.UserCodeConfig{
				CharSet:      "BCDFGHJKLMNPQRSTVWXZ0123456789", // No vowels, avoid words
				CharAmount:   8,
				DashInterval: 4, // Format: XXXX-XXXX
			},
		},
	}
	issuer := func(_ bool) (fn op.IssuerFromRequest, err error) {
		fn = Issuer
		return
	}
	builtin.provider, err = op.NewProvider(
		config,
		builtin.storage,
		issuer,
		op.WithAllowInsecure(),
		op.WithCustomTokenEndpoint(op.NewEndpoint("token")),
		op.WithCustomIntrospectionEndpoint(op.NewEndpoint("introspect")),
		op.WithCustomEndSessionEndpoint(op.NewEndpoint("logout")),
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	builtin.idpHandler = &FedIdpHandler{
		db:      db,
		storage: builtin.storage,
		cache:   cache,
	}
	builtin.storage.idpHandler = builtin.idpHandler
	ds := LDAP{
		Kind:        federated.Ldap.Kind,
		Name:        federated.Ldap.Name,
		URL:         federated.Ldap.URL,
		BaseDN:      federated.Ldap.BaseDN,
		BindDN:      federated.Ldap.BindDN,
		Password:    federated.Ldap.Password,
		UserFilter:  federated.Ldap.UserFilter,
		GroupFilter: federated.Ldap.GroupFilter,
		HasMemberOf: federated.Ldap.HasMemberOf,
		TLS:         federated.Ldap.TLS,
	}
	ds.mapper.Use(federated.Ldap.RoleMappings)
	builtin.dsHandler = &LdapHandler{
		enabled: federated.Ldap.Enabled,
		cache:   builtin.cache,
		db:      db,
		ds:      ds,
	}
	builtin.storage.dsHandler = builtin.dsHandler
	builtin.dagHandler = &DagHandler{
		storage: builtin.storage,
	}
	return
}

// Builtin provides OIDC authentication.
type Builtin struct {
	db         *gorm.DB
	provider   op.OpenIDProvider
	idpHandler *FedIdpHandler
	dsHandler  *LdapHandler
	dagHandler *DagHandler
	cache      *Cache
	storage    *Storage
	keySet     KeySet
}

// Handler returns an http handler.
func (p *Builtin) Handler() (h http.Handler) {
	h = p.provider
	return
}

// DagHandler returns the device authorization grant handler.
func (p *Builtin) DagHandler() (h *DagHandler) {
	h = p.dagHandler
	return
}

// IdpHandler returns the IdP federation handler.
func (p *Builtin) IdpHandler() (h *FedIdpHandler) {
	h = p.idpHandler
	return
}

// Cache returns the provider cache.
func (p *Builtin) Cache() *Cache {
	return p.cache
}

// Login handles the custom login page.
func (p *Builtin) Login(
	writer http.ResponseWriter,
	request *http.Request,
	authReqId string) (err error) {
	err = p.storage.Login(writer, request, authReqId)
	return
}

// NewToken creates a new personal access token.
func (p *Builtin) NewToken(subject string, lifespan time.Duration) (m Token, err error) {
	defer func() {
		if err == nil {
			err = p.db.Save(&m).Error
			if err != nil {
				err = liberr.Wrap(err)
			}
			if err == nil {
				p.cache.TokenSaved(&m)
			}
		}
	}()
	m = p.newToken(lifespan)
	s, err := p.cache.FindSubject(subject)
	if err != nil {
		return
	}
	if s.IsUser() {
		m.UserID = s.UserId
	}
	if s.IsIdentity() {
		m.IdpIdentityID = s.IdentityId
	}
	if s.IsClient() {
		m.IdpClientID = s.ClientId
	}
	return
}

// TaskGrant creates a new task api-key.
func (p *Builtin) TaskGrant(taskId uint) (m Token, err error) {
	m = p.newToken(0)
	m.TaskID = &taskId
	err = p.db.Create(&m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	p.cache.TokenSaved(&m)
	p.cache.TaskGranted(taskId)
	return
}

// TaskRevoke revokes a task token.
func (p *Builtin) TaskRevoke(taskId uint) {
	p.cache.TaskRevoked(taskId)
	err := p.db.Where("TaskID", taskId).Delete(&Token{}).Error
	if err != nil {
		Log.Error(err,
			"Task revoke failed.",
			"taskId",
			taskId)
	}
	return
}

// Authenticate authenticates the user making the web request.
func (p *Builtin) Authenticate(req *Request) (jwToken *jwt.Token, err error) {
	defer func() {
		if err == nil {
			return
		}
		if !errors.Is(err, &NotAuthenticated{}) {
			err = &NotAuthenticated{
				Reason: err.Error(),
			}
		}
		Log.V(2).Info(
			"Authentication failed.",
			"reason",
			err.Error())
	}()
	if req.Token != "" {
		jwToken, err = p.authToken(req)
		return
	}
	if req.Login != "" {
		for _, method := range []func(*Request) (*jwt.Token, error){
			p.authUser,
			p.authLdapUser,
		} {
			jwToken, err = method(req)
			if errors.Is(err, &NotFound{}) {
				continue
			} else {
				break
			}
		}
		return
	}
	//
	err = liberr.Wrap(
		&NotAuthenticated{
			Reason: "missing credentials",
		})
	return
}

// Revoke revokes a token by ID.
func (p *Builtin) Revoke(tokenId uint) (err error) {
	p.cache.TokenDeleted(tokenId)
	m := &Token{}
	m.ID = tokenId
	err = p.db.Delete(m).Error
	return
}

// User returns the username from the JWT token.
func (p *Builtin) User(jwToken *jwt.Token) (user string) {
	s := p.Subject(jwToken)
	subject, err := p.cache.FindSubject(s)
	if err == nil {
		user = subject.Login()
	}
	return
}

// Subject returns the subject from the JWT token.
func (p *Builtin) Subject(jwToken *jwt.Token) (subject string) {
	claims := jwToken.Claims.(jwt.MapClaims)
	v := claims[ClaimSub]
	if s, cast := v.(string); cast {
		subject = s
	}
	return
}

// Scopes returns a list of scopes.
func (p *Builtin) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	claims := jwToken.Claims.(jwt.MapClaims)
	v := claims[ClaimScope]
	if sList, cast := v.(string); cast {
		for _, s := range strings.Fields(sList) {
			scope := &BaseScope{}
			scope.With(s)
			scopes = append(
				scopes,
				scope)
		}
	}
	return
}

// validToken returns an error if not valid.
func (p *Builtin) validToken(jwToken *jwt.Token, req *Request) (err error) {
	if !jwToken.Valid {
		err = liberr.Wrap(&NotAuthenticated{
			Reason: "token invalid",
			Token:  jwToken.Raw,
		})
		return
	}
	claims, cast := jwToken.Claims.(jwt.MapClaims)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Claims not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	v, found := claims[ClaimSub]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not string.",
				Token:  jwToken.Raw,
			})
		return
	}
	v, found = claims[ClaimScope]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not string.",
				Token:  jwToken.Raw,
			})
		return
	}
	v, found = claims[ClaimExp]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Exp not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	f64, cast := v.(float64)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Exp not float64.",
				Token:  jwToken.Raw,
			})
		return
	}
	expiration := time.Unix(int64(f64), 0)
	if expiration.Before(time.Now()) {
		err = &NotValid{
			Reason: "Token expired.",
			Token:  jwToken.Raw,
		}
		return
	}
	v, found = claims[ClaimIss]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Iss not specified.",
				Token:  jwToken.Raw,
			})
		return
	}
	issuerStr, cast := v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Iss not string.",
				Token:  jwToken.Raw,
			})
		return
	}
	if issuerStr != Issuer(req.CTX.Request) {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Iss mismatch.",
				Token:  jwToken.Raw,
			})
		return
	}
	return
}

// KeySet represents a JSON Web Key Set.
type KeySet struct {
	Keys []JWK
}

// SigningKey returns the primary signing key.
func (k *KeySet) SigningKey() (key op.SigningKey) {
	if len(k.Keys) > 0 {
		key = &k.Keys[0]
	}
	return
}

// Key returns a key by ID.
func (k *KeySet) Key(id string) (jwk JWK, err error) {
	for _, key := range k.Keys {
		if key.KeyID == id {
			jwk = key
			return
		}
	}
	err = errors.New("key not found")
	return
}

// authToken authenticates the token.
func (p *Builtin) authToken(req *Request) (jwToken *jwt.Token, err error) {
	defer func() {
		if errors.Is(err, &NotValid{}) {
			Log.V(2).Info("[builtin] " + err.Error())
		}
	}()
	token := req.Token
	//
	// RSA Token.
	jwToken, err = jwt.Parse(
		token,
		func(jwToken *jwt.Token) (key any, err error) {
			_, cast := jwToken.Method.(*jwt.SigningMethodRSA)
			if !cast {
				err = liberr.Wrap(&NotAuthenticated{
					Reason: "invalid token signing method",
					Token:  token,
				})
				return
			}
			kid, found := jwToken.Header["kid"]
			if !found {
				err = liberr.Wrap(&NotAuthenticated{
					Reason: "missing key ID",
					Token:  token,
				})
				return
			}
			jwk, findErr := p.keySet.Key(kid.(string))
			if findErr != nil {
				err = liberr.Wrap(findErr)
				return
			}
			privateKey, cast := jwk.Key().(*rsa.PrivateKey)
			if !cast {
				err = liberr.Wrap(&NotAuthenticated{
					Reason: "invalid key type",
					Token:  token,
				})
				return
			}
			key = &privateKey.PublicKey
			return
		},
		jwt.WithoutClaimsValidation())
	//
	// Legacy HMAC token.
	if err != nil {
		jwToken, err = jwt.Parse(
			token,
			func(jwToken *jwt.Token) (secret any, err error) {
				_, cast := jwToken.Method.(*jwt.SigningMethodHMAC)
				if !cast {
					err = liberr.Wrap(&NotAuthenticated{
						Reason: "invalid token signing method",
						Token:  token,
					})
					return
				}
				secret = []byte(Settings.Token.Key)
				claims, cast := jwToken.Claims.(jwt.MapClaims)
				if cast {
					claims[ClaimIss] = Issuer(req.CTX.Request)
				}
				return
			},
			jwt.WithoutClaimsValidation())
	}
	if err == nil {
		err = p.validToken(jwToken, req)
		return
	}
	//
	// PAT/api-key.
	pat, err := p.cache.FindToken(req.Token)
	if err == nil {
		jwToken = jwt.New(jwt.SigningMethodHS512)
		jwtClaims := jwToken.Claims.(jwt.MapClaims)
		jwtClaims[ClaimSub] = pat.Subject
		jwtClaims[ClaimScope] = pat.Scopes
		jwtClaims[ClaimIss] = Issuer(req.CTX.Request)
		jwtClaims[ClaimIat] = time.Now().Unix()
		jwtClaims[ClaimExp] = pat.Expiration.Unix()
		return
	}
	err = liberr.Wrap(&NotAuthenticated{
		Reason: "token not recognized",
		Token:  token,
	})
	return
}

// authUser authenticates the user.
func (p *Builtin) authUser(req *Request) (jwToken *jwt.Token, err error) {
	login := req.Login
	password := req.Password
	user, err := p.cache.FindUserByLogin(login)
	if err != nil {
		return
	}
	if !secret.MatchPassword(password, user.Password) {
		err = &NotAuthenticated{
			Reason: "invalid password",
			Token:  login,
		}
		return
	}
	scopes := make([]string, 0)
	for _, ref := range user.Roles {
		role, nErr := p.cache.FindRoleById(ref.ID)
		if nErr != nil {
			// not found.
			continue
		}
		for _, scope := range role.GetScopes() {
			scopes = append(scopes, scope)
		}
	}
	scopes = uniqueStrings(scopes)
	sort.Strings(scopes)
	jwToken = jwt.New(jwt.SigningMethodRS256)
	jwtClaims := jwToken.Claims.(jwt.MapClaims)
	jwtClaims[ClaimSub] = user.Subject
	jwtClaims[ClaimScope] = strings.Join(scopes, " ")
	jwtClaims[ClaimIss] = Issuer(req.CTX.Request)
	jwtClaims[ClaimIat] = time.Now().Unix()
	jwtClaims[ClaimExp] = time.Now().Add(Settings.Auth.BasicAuthLifespan).Unix()
	return
}

// authLdapUser authenticates an LDAP user.
func (p *Builtin) authLdapUser(req *Request) (jwToken *jwt.Token, err error) {
	lifespan := Settings.Auth.BasicAuthLifespan
	subject, err := p.dsHandler.Authenticate(req.Login, req.Password, lifespan)
	if err != nil {
		return
	}
	jwToken = jwt.New(jwt.SigningMethodRS256)
	jwtClaims := jwToken.Claims.(jwt.MapClaims)
	jwtClaims[ClaimSub] = subject.Key
	jwtClaims[ClaimScope] = strings.Join(subject.Scopes, " ")
	jwtClaims[ClaimIss] = Issuer(req.CTX.Request)
	jwtClaims[ClaimIat] = time.Now().Unix()
	jwtClaims[ClaimExp] = time.Now().Add(lifespan).Unix()
	return
}

// newToken returns a new token.
func (p *Builtin) newToken(lifespan time.Duration) (token Token) {
	if lifespan == 0 {
		lifespan = Settings.APIKey.Lifespan
	}
	token.Kind = KindAPIKey
	token.AuthId = p.storage.genId()
	token.Secret = p.storage.genId()
	token.Digest = secret.Hash(token.Secret)
	token.Expiration = time.Now().Add(lifespan)
	return
}

// JWK is a JSON Web Key.
type JWK struct {
	KeyID      string
	Algorithm  string
	Use        string
	PrivateKey any
}

// SignatureAlgorithm returns the signature algorithm.
func (j *JWK) SignatureAlgorithm() (s jose.SignatureAlgorithm) {
	s = jose.SignatureAlgorithm(j.Algorithm)
	return
}

// Key returns the private key.
func (j *JWK) Key() any {
	return j.PrivateKey
}

// ID returns the key ID.
func (j *JWK) ID() (s string) {
	s = j.KeyID
	return
}
