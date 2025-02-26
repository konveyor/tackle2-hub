package auth

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	liberr "github.com/jortel/go-utils/error"
	"gorm.io/gorm"
)

// Validators provide token validation based on claims.
var Validators []Validator

// Validator provides token validation.
type Validator interface {
	// Valid determines if the token is valid.
	// When valid, return nil.
	// When not valid, return NotValid error.
	// On failure, return the (cause) error.
	Valid(token *jwt.Token, db *gorm.DB) (err error)
}

// NoAuth provider always permits access.
type NoAuth struct {
}

// NewToken creates a new signed token.
func (r NoAuth) NewToken(user string, scopes []string, claims jwt.MapClaims) (signed string, err error) {
	return
}

// Authenticate the token
func (r *NoAuth) Authenticate(_ *Request) (jwToken *jwt.Token, err error) {
	return
}

// Scopes decodes a list of scopes from the token.
// For the NoAuth provider, this just returns a single
// wildcard scope matching everything.
func (r *NoAuth) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	scopes = append(scopes, &BaseScope{"*", "*"})
	return
}

// User mocks username for NoAuth
func (r *NoAuth) User(jwToken *jwt.Token) (name string) {
	name = "admin.noauth"
	return
}

// Login and obtain a token.
func (r *NoAuth) Login(user, password string) (token Token, err error) {
	return
}

// Refresh token.
func (r *NoAuth) Refresh(refresh string) (token Token, err error) {
	return
}

// Builtin auth provider.
type Builtin struct {
}

// Authenticate the token
func (r *Builtin) Authenticate(request *Request) (jwToken *jwt.Token, err error) {
	defer func() {
		if errors.Is(err, &NotValid{}) {
			Log.V(2).Info("[builtin] " + err.Error())
		}
	}()
	token, err := r.parseToken(request)
	if err != nil {
		return
	}
	jwToken, err = jwt.Parse(
		token,
		func(jwToken *jwt.Token) (secret any, err error) {
			_, cast := jwToken.Method.(*jwt.SigningMethodHMAC)
			if !cast {
				err = liberr.Wrap(&NotAuthenticated{Token: token})
				return
			}
			secret = []byte(Settings.Auth.Token.Key)
			return
		})
	if err != nil {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	if !jwToken.Valid {
		err = liberr.Wrap(&NotAuthenticated{Token: token})
		return
	}
	claims, cast := jwToken.Claims.(jwt.MapClaims)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Claims not specified.",
				Token:  token,
			})
		return
	}
	v, found := claims["user"]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not specified.",
				Token:  token,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "User not string.",
				Token:  token,
			})
		return
	}
	v, found = claims["scope"]
	if !found {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not specified.",
				Token:  token,
			})
		return
	}
	_, cast = v.(string)
	if !cast {
		err = liberr.Wrap(
			&NotValid{
				Reason: "Scope not string.",
				Token:  token,
			})
		return
	}
	for _, v := range Validators {
		err = v.Valid(jwToken, request.DB)
		if err != nil {
			return
		}
	}
	return
}

// Scopes returns a list of scopes.
func (r *Builtin) Scopes(jwToken *jwt.Token) (scopes []Scope) {
	claims := jwToken.Claims.(jwt.MapClaims)
	for _, s := range strings.Fields(claims["scope"].(string)) {
		scope := &BaseScope{}
		scope.With(s)
		scopes = append(
			scopes,
			scope)
	}
	return
}

// User returns the user associated with the token.
func (r *Builtin) User(jwToken *jwt.Token) (user string) {
	claims := jwToken.Claims.(jwt.MapClaims)
	user = claims["user"].(string)
	return
}

// Login and obtain a token.
func (r *Builtin) Login(user, password string) (token Token, err error) {
	return
}

// Refresh token.
func (r *Builtin) Refresh(refresh string) (token Token, err error) {
	return
}

// NewToken creates a new signed token.
func (r *Builtin) NewToken(user string, scopes []string, claims jwt.MapClaims) (signed string, err error) {
	token := jwt.New(jwt.SigningMethodHS512)
	jwtClaims := token.Claims.(jwt.MapClaims)
	for k, v := range claims {
		jwtClaims[k] = v
	}
	jwtClaims["user"] = user
	jwtClaims["scope"] = strings.Join(scopes, " ")
	signed, err = token.SignedString([]byte(Settings.Auth.Token.Key))
	return
}

// parseToken returns the token
func (r *Builtin) parseToken(request *Request) (token string, err error) {
	splitToken := strings.Fields(request.Token)
	if len(splitToken) != 2 || strings.ToLower(splitToken[0]) != "bearer" {
		err = liberr.Wrap(&NotValid{Token: request.Token})
		return
	}
	token = splitToken[1]
	return
}
