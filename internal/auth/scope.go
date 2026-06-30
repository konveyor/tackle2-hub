package auth

import "strings"

var verbs = []string{
	"decrypt",
	"delete",
	"get",
	"patch",
	"post",
	"put",
}

// Scope provides scope behavior.
type Scope struct {
	Resource string
	Method   string
}

// With parses a scope and populate fields.
// Format: <resource>:<method>
func (r *Scope) With(s string) {
	part := strings.Split(s, ":")
	n := len(part)
	if n > 0 {
		r.Resource = part[0]
	}
	if n > 1 {
		r.Method = part[1]
	}
	return
}

// Match returns whether the scope is a match.
func (r *Scope) Match(resource string, method string) (b bool) {
	b = (r.Resource == "*" ||
		r.Eq(r.Resource, resource)) &&
		(r.Method == "*" ||
			r.Eq(r.Method, method))
	return
}

// Expand returns scopes with expanded wildcards.
func (r *Scope) Expand() (expanded []Scope) {
	var nouns []string
	var methods []string
	if r.Resource == "*" {
		for n := range registeredResources {
			nouns = append(nouns, n)
		}
	} else {
		nouns = []string{r.Resource}
	}
	if r.Method == "*" {
		methods = verbs[:]
	} else {
		methods = []string{r.Method}
	}
	for _, n := range nouns {
		for _, m := range methods {
			expanded = append(expanded, Scope{Resource: n, Method: m})
		}
	}
	return
}

// Eq returns true when strings matched.
func (r *Scope) Eq(a, b string) (eq bool) {
	eq = strings.EqualFold(a, b)
	return
}

// String representations of the scope.
func (r *Scope) String() (s string) {
	s = strings.Join([]string{r.Resource, r.Method}, ":")
	return
}

// ExpandScopes returns the scopes with wildcards expanded.
// Uses Scope.Expand().
func ExpandScopes(in ...string) (out []string) {
	for _, n := range in {
		s := Scope{}
		s.With(n)
		for _, s = range s.Expand() {
			out = append(out, s.String())
		}
	}
	return
}
