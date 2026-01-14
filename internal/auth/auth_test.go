package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/onsi/gomega"
)

type _TestProvider struct {
	err error
	NoAuth
}

func (p *_TestProvider) Authenticate(_ *Request) (jwToken *jwt.Token, err error) {
	err = p.err
	return
}

func TestValid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	p := Builtin{}
	user := "myUser"
	scopes := []string{
		"things:get",
		"things:post",
	}
	//
	// New token.
	signed, err := p.NewToken(user, scopes, nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(signed) > 0).To(gomega.BeTrue())
	//
	// Authenticate
	jwToken, err := p.Authenticate(&Request{Token: "Bearer " + signed})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(jwToken != nil).To(gomega.BeTrue())
	//
	// Scopes.
	tokenScopes := p.Scopes(jwToken)
	g.Expect(tokenScopes).To(
		gomega.Equal([]Scope{
			&BaseScope{Resource: "things", Method: "get"},
			&BaseScope{Resource: "things", Method: "post"},
		}))
	//
	// User
	tokenUser := p.User(jwToken)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(tokenUser).To(gomega.Equal(tokenUser))
}

func TestNotValid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	p := Builtin{}
	user := "myUser"
	scopes := []string{
		"things:get",
		"things:post",
	}
	//
	// New token.
	signed, err := p.NewToken(user, scopes, nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(signed) > 0).To(gomega.BeTrue())
	//
	// Valid
	jwToken, err := p.Authenticate(&Request{Token: "Bearer " + signed})
	g.Expect(err).To(gomega.BeNil())
	g.Expect(jwToken != nil).To(gomega.BeTrue())
	//
	// Not valid.
	Settings.Auth.Token.Key = "NotMyKey"
	jwToken, err = p.Authenticate(&Request{Token: "Bearer " + signed})
	g.Expect(err != nil).To(gomega.BeTrue())
}

func TestScope(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	scope := BaseScope{}
	//
	// Parsing & match.
	scope.With("things:get")
	g.Expect(scope.Resource).To(gomega.Equal("things"))
	g.Expect(scope.Method).To(gomega.Equal("get"))
	g.Expect(scope.Match("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Match("Things", "Get")).To(gomega.BeTrue())
	g.Expect(scope.Match("xx", "get")).To(gomega.BeFalse())
	g.Expect(scope.Match("things", "xx")).To(gomega.BeFalse())
	scope.With("things")
	g.Expect(scope.Resource).To(gomega.Equal("things"))

	//
	// Wildcard.
	scope.With("things:*")
	g.Expect(scope.Match("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Match("things", "xx")).To(gomega.BeTrue())
	g.Expect(scope.Match("xx", "get")).To(gomega.BeFalse())
	//
	// Wildcard.
	scope.With("*:get")
	g.Expect(scope.Match("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Match("things", "xx")).To(gomega.BeFalse())
	g.Expect(scope.Match("xx", "get")).To(gomega.BeTrue())
	//
	// Wildcard.
	scope.With("*:*")
	g.Expect(scope.Match("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Match("xx", "xx")).To(gomega.BeTrue())
}

func TestRequestHubPermit(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	Hub = &Builtin{}
	Remote = &_TestProvider{err: &NotAuthenticated{}}
	user := "myUser"
	scopes := []string{
		"things:get",
		"things:post",
	}
	//
	// New token.
	signed, err := Hub.NewToken(user, scopes, nil)
	g.Expect(err).To(gomega.BeNil())
	//
	// Permit
	request := Request{
		Token:  "Bearer " + signed,
		Scope:  "things",
		Method: "GET",
	}
	result, err := request.Permit()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Authenticated).To(gomega.BeTrue())
	g.Expect(result.User).To(gomega.Equal(user))
	g.Expect(len(result.Scopes)).To(gomega.Equal(len(scopes)))
	for i := range scopes {
		g.Expect(result.Scopes[i].String()).To(gomega.Equal(scopes[i]))
	}
}

func TestRequestRemotePermit(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	Hub = &_TestProvider{err: &NotAuthenticated{}}
	Remote = &Builtin{}
	user := "myUser"
	scopes := []string{
		"things:get",
		"things:post",
	}
	//
	// New token.
	signed, err := Remote.NewToken(user, scopes, nil)
	g.Expect(err).To(gomega.BeNil())
	//
	// Permit
	request := Request{
		Token:  "Bearer " + signed,
		Scope:  "things",
		Method: "GET",
	}
	result, err := request.Permit()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Authenticated).To(gomega.BeTrue())
	g.Expect(result.User).To(gomega.Equal(user))
	g.Expect(len(result.Scopes)).To(gomega.Equal(len(scopes)))
	for i := range scopes {
		g.Expect(result.Scopes[i].String()).To(gomega.Equal(scopes[i]))
	}
}

func TestRequestPermitNotAuthenticated(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	Hub = &_TestProvider{err: &NotAuthenticated{}}
	Remote = &_TestProvider{err: &NotAuthenticated{}}
	//
	// Permit
	request := Request{
		Token:  "",
		Scope:  "things",
		Method: "PUT",
	}
	result, err := request.Permit()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Authenticated).To(gomega.BeFalse())
	g.Expect(result.Authorized).To(gomega.BeFalse())
	g.Expect(result.User).To(gomega.Equal(""))
	g.Expect(len(result.Scopes)).To(gomega.Equal(0))
}

func TestRequestPermitNotNotValid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	Hub = &Builtin{}
	Remote = &_TestProvider{err: &NotValid{}}
	//
	// Permit
	request := Request{
		Token:  "ABCD",
		Scope:  "things",
		Method: "PUT",
	}
	result, err := request.Permit()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Authenticated).To(gomega.BeFalse())
	g.Expect(result.Authorized).To(gomega.BeFalse())
	g.Expect(result.User).To(gomega.Equal(""))
	g.Expect(len(result.Scopes)).To(gomega.Equal(0))
}

func TestRequestHubPermitNotAuthorized(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	Hub = &Builtin{}
	Remote = &_TestProvider{err: &NotAuthenticated{}}
	user := "myUser"
	scopes := []string{
		"things:get",
		"things:post",
	}
	//
	// New token.
	signed, err := Hub.NewToken(user, scopes, nil)
	g.Expect(err).To(gomega.BeNil())
	//
	// Permit
	request := Request{
		Token:  "Bearer " + signed,
		Scope:  "things",
		Method: "PUT",
	}
	result, err := request.Permit()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Authenticated).To(gomega.BeTrue())
	g.Expect(result.Authorized).To(gomega.BeFalse())
	g.Expect(result.User).To(gomega.Equal(""))
	g.Expect(len(result.Scopes)).To(gomega.Equal(0))
}

func TestRequestRemotePermitNotAuthorized(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	Settings.Auth.Token.Key = "TestKey"
	Remote = &Builtin{}
	Hub = &_TestProvider{err: &NotAuthenticated{}}
	user := "myUser"
	scopes := []string{
		"things:get",
		"things:post",
	}
	//
	// New token.
	signed, err := Remote.NewToken(user, scopes, nil)
	g.Expect(err).To(gomega.BeNil())
	//
	// Permit
	request := Request{
		Token:  "Bearer " + signed,
		Scope:  "things",
		Method: "PUT",
	}
	result, err := request.Permit()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Authenticated).To(gomega.BeTrue())
	g.Expect(result.Authorized).To(gomega.BeFalse())
	g.Expect(result.User).To(gomega.Equal(""))
	g.Expect(len(result.Scopes)).To(gomega.Equal(0))
}
