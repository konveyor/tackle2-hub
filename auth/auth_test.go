package auth

import (
	"github.com/onsi/gomega"
	"testing"
)

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
	jwToken, err := p.Authenticate(signed)
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
	jwToken, err := p.Authenticate(signed)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(jwToken != nil).To(gomega.BeTrue())
	//
	// Not valid.
	Settings.Auth.Token.Key = "NotMyKey"
	jwToken, err = p.Authenticate(signed)
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
