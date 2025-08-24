// Package cmp provides object comparison.
// Example Eq() report:
//
// Expected:
// __________________________
// id: 1
// list:
// - 1
// - 3
// - 2
// thing:
//   name: Roger
//   age: 18
//   list:
//   - 1
//   - 2
//   - 3
//   - 4
//   - 5
//
// Got:
// __________________________
// id: 1
// list:
// - 1
// - 2
// - 3
// - 4
// - 5
// thing:
//   name: Brian
//   age: 18
//   list: []
//
// Diff
// __________________________
// ~ List[1]=2 expected: 3
// ~ List[2]=3 expected: 2
// + List[3]=4
// + List[4]=5
// ~ Thing.Name=Brian expected: Roger
// - Thing.List[0]=1
// - Thing.List[1]=2
// - Thing.List[2]=3
// - Thing.List[3]=4
// - Thing.List[4]=5
package cmp

import (
	"gopkg.in/yaml.v2"
)

func New(ignoredPaths ...string) (cmp *Differ) {
	return &Differ{IgnoredPaths: ignoredPaths}
}

// Pretty returns a YAML representation.
func Pretty(a any) (s string) {
	b, err := yaml.Marshal(a)
	if err != nil {
		panic(err)
	}
	s = string(b)
	return
}

func Eq(expected, got any, ignoredPaths ...string) (eq bool, report string) {
	cmp := New(ignoredPaths...)
	eq, report = cmp.Eq(expected, got)
	return
}

func Inspect(a, b any, ignoredPaths ...string) (eq bool, d string) {
	cmp := New(ignoredPaths...)
	eq, d = cmp.Inspect(a, b)
	return
}
