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
			&BaseScope{resource: "things", method: "get"},
			&BaseScope{resource: "things", method: "post"},
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
	// Parsing.
	scope.With("things:get:a:b")
	g.Expect(scope.resource).To(gomega.Equal("things"))
	g.Expect(scope.method).To(gomega.Equal("get"))
	g.Expect(scope.mods).To(gomega.Equal([]string{"a", "b"}))
	g.Expect(scope.Allow("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Allow("xx", "get")).To(gomega.BeFalse())
	g.Expect(scope.Allow("things", "xx")).To(gomega.BeFalse())
	//
	// Wildcard.
	scope.With("things:*")
	g.Expect(scope.Allow("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Allow("things", "xx")).To(gomega.BeTrue())
	g.Expect(scope.Allow("xx", "get")).To(gomega.BeFalse())
	//
	// Wildcard.
	scope.With("*:get")
	g.Expect(scope.Allow("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Allow("things", "xx")).To(gomega.BeFalse())
	g.Expect(scope.Allow("xx", "get")).To(gomega.BeTrue())
	//
	// Wildcard.
	scope.With("*:*")
	g.Expect(scope.Allow("things", "get")).To(gomega.BeTrue())
	g.Expect(scope.Allow("xx", "xx")).To(gomega.BeTrue())
}
