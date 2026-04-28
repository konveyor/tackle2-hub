package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/op"
	"gorm.io/gorm"
)

// NewBuiltin returns a configured provider.
func NewBuiltin(db *gorm.DB) (builtin *Builtin, err error) {
	cache := NewCache(db)
	err = cache.Refresh()
	if err != nil {
		return
	}
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
		db:            db,
		keySet:        builtin.keySet,
		authReqs:      make(map[string]*AuthRequest),
		authByCode:    make(map[string]string),
		devAuthReqs:   make(map[string]*DeviceAuthRequest),
		devAuthByCode: make(map[string]string),
		clientById:    make(map[string]op.Client),
		cache:         cache,
	}
	for _, client := range Settings.Clients {
		opClient := &Client{}
		opClient.With(&client)
		builtin.storage.clientById[client.Id] = opClient
	}
	builtin.storage.clientById[DevVerifierClientId] =
		&Client{
			id:              DevVerifierClientId,
			applicationType: op.ApplicationTypeWeb,
			grantTypes:      []string{"authorization_code"},
			redirectURIs:    []string{Settings.Auth.IssuerWithPath(api.AuthDevAuthCallback)},
			scopes:          []string{"openid"},
		}
	issuer := Settings.IssuerURL
	config := &op.Config{
		CodeMethodS256:          true,
		AuthMethodPost:          true,
		AuthMethodPrivateKeyJWT: false,
		GrantTypeRefreshToken:   true,
		RequestObjectSupported:  false,
		DeviceAuthorization: op.DeviceAuthorizationConfig{
			Lifetime:     15 * time.Minute,
			PollInterval: 5 * time.Second,
			UserFormPath: "/device",
			UserCode: op.UserCodeConfig{
				CharSet:      "BCDFGHJKLMNPQRSTVWXZ0123456789", // No vowels to avoid words
				CharAmount:   8,
				DashInterval: 4, // Format: XXXX-XXXX
			},
		},
	}
	builtin.provider, err = op.NewProvider(
		config,
		builtin.storage,
		op.StaticIssuer(issuer),
		op.WithAllowInsecure(),
		op.WithCustomTokenEndpoint(op.NewEndpoint("token")),
		op.WithCustomIntrospectionEndpoint(op.NewEndpoint("introspect")),
		op.WithCustomEndSessionEndpoint(op.NewEndpoint("logout")),
	)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	// Initialize RP client and IdpHandler if external IdP is enabled
	Log.Info("Federation configuration",
		"enabled", federation.Enabled,
		"issuer", federation.Idp.Issuer,
		"clientId", federation.Idp.ClientId,
		"redirectURI", federation.Idp.RedirectURI)
	if federation.Enabled {
		Log.Info("Initializing IdP federation handler")
		var rpClient rp.RelyingParty
		rpClient, err = rp.NewRelyingPartyOIDC(
			context.Background(),
			federation.Idp.Issuer,
			federation.Idp.ClientId,
			federation.Idp.ClientSecret,
			federation.Idp.RedirectURI,
			federation.Idp.Scopes,
		)
		if err != nil {
			Log.Error(err, "Failed to create RP client for IdP federation")
			err = liberr.Wrap(err)
			return
		}
		builtin.idpHandler = &IdpHandler{
			rpClient: rpClient,
			db:       db,
			storage:  builtin.storage,
			cache:    cache,
		}
		builtin.storage.idpHandler = builtin.idpHandler
		Log.Info("IdP federation handler initialized successfully")
	} else {
		Log.Info("IdP federation disabled")
	}
	// Initialize DagHandler
	builtin.dagHandler = &DagHandler{
		storage: builtin.storage,
	}
	return
}

// Builtin provides OIDC authentication.
type Builtin struct {
	db         *gorm.DB
	provider   op.OpenIDProvider
	idpHandler *IdpHandler
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
func (p *Builtin) IdpHandler() (h *IdpHandler) {
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

// NewPAT create a new personal access token.
func (p *Builtin) NewPAT(subject string, lifespan time.Duration) (m Token, err error) {
	defer func() {
		if err != nil {
			err = liberr.Wrap(err)
		}
	}()
	defer func() {
		if err == nil {
			err = p.db.Save(&m).Error
			if err == nil {
				p.cache.TokenSaved(&m)
			}
		}
	}()
	m = p.newToken(lifespan)
	s, err := p.cache.FindSubject(subject)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if s.IsUser() {
		m.UserID = s.userId
	}
	if s.IsIdentity() {
		m.IdpIdentityID = s.identityId
	}
	return
}

// NewTaskToken create a new task api-key.
func (p *Builtin) NewTaskToken(taskId uint) (m Token, err error) {
	m = p.newToken(0)
	task, err := p.cache.GetTask(taskId)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	m.TaskID = &task.ID
	err = p.db.Create(&m).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	p.cache.TokenSaved(&m)
	return
}

// Authenticate the user making the web request.
func (p *Builtin) Authenticate(req *Request) (jwToken *jwt.Token, err error) {
	defer func() {
		if errors.Is(err, &NotValid{}) {
			Log.V(2).Info("[builtin] " + err.Error())
		}
	}()
	if req.Token != "" {
		jwToken, err = p.authToken(req)
		return
	}
	if req.Userid != "" {
		jwToken, err = p.authUser(req)
		return
	}
	err = liberr.Wrap(&NotAuthenticated{})
	return
}

// Revoke a token.
func (p *Builtin) Revoke(tokenId uint) (err error) {
	p.cache.TokenDeleted(tokenId)
	m := &model.Token{}
	m.ID = tokenId
	err = p.db.Delete(m).Error
	return
}

func (r *Builtin) User(jwToken *jwt.Token) (user string) {
	claims := jwToken.Claims.(jwt.MapClaims)
	v := claims[ClaimSub]
	if s, cast := v.(string); cast {
		user = s
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
func (p *Builtin) validToken(jwToken *jwt.Token) (err error) {
	if !jwToken.Valid {
		err = liberr.Wrap(&NotAuthenticated{Token: jwToken.Raw})
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

// authToken authenticate the token.
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
				err = liberr.Wrap(&NotAuthenticated{Token: token})
				return
			}
			kid, found := jwToken.Header["kid"]
			if !found {
				err = liberr.Wrap(&NotAuthenticated{Token: token})
				return
			}
			jwk, findErr := p.keySet.Key(kid.(string))
			if findErr != nil {
				err = liberr.Wrap(findErr)
				return
			}
			privateKey, cast := jwk.Key().(*rsa.PrivateKey)
			if !cast {
				err = liberr.Wrap(&NotAuthenticated{Token: token})
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
					err = liberr.Wrap(&NotAuthenticated{Token: token})
					return
				}
				secret = []byte(Settings.Token.Key)
				return
			},
			jwt.WithoutClaimsValidation())
	}
	if err == nil {
		err = p.validToken(jwToken)
		return
	}
	//
	// PAT/api-key.
	pat, err := p.cache.GetToken(req.Token)
	if err == nil {
		jwToken = jwt.New(jwt.SigningMethodHS512)
		jwtClaims := jwToken.Claims.(jwt.MapClaims)
		jwtClaims[ClaimScope] = pat.Scopes
		jwtClaims[ClaimSub] = pat.Subject
		return
	}
	err = liberr.Wrap(&NotAuthenticated{Token: token})
	return
}

// authUser authenticate the user.
func (p *Builtin) authUser(req *Request) (jwToken *jwt.Token, err error) {
	userid := req.Userid
	password := req.Password
	user, err := p.cache.FindUserByUserid(userid)
	if err != nil {
		if errors.Is(err, &NotFound{}) {
			err = &NotAuthenticated{
				Token: userid,
			}
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	if !secret.MatchPassword(password, user.Password) {
		err = &NotAuthenticated{
			Token: userid,
		}
		return
	}
	unique := make(map[string]bool)
	scopes := make([]string, 0)
	for _, u := range user.Roles {
		for _, perm := range u.Permissions {
			if !unique[perm.Scope] {
				scopes = append(scopes, perm.Scope)
			}
			unique[perm.Scope] = true
		}
	}
	jwToken = jwt.New(jwt.SigningMethodRS256)
	jwtClaims := jwToken.Claims.(jwt.MapClaims)
	jwtClaims[ClaimSub] = user.Subject
	jwtClaims[ClaimScope] = strings.Join(scopes, " ")
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

// JWK a JSON Web Key.
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
