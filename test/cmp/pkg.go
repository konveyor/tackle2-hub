// Package cmp provides object comparison.
// Example Cmp() report:
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
	"strings"

	"github.com/davecgh/go-spew/spew"
)

func New(ignoredPaths ...string) (cmp *Cmp) {
	cmp = &Cmp{}
	for _, path := range ignoredPaths {
		part := []string{}
		for _, p := range strings.Split(path, ".") {
			if len(part) > 0 {
				p = "." + p
			}
			part = append(part, p)
		}
		cmp.IgnoredPaths = append(cmp.IgnoredPaths, part)
	}
	return
}

// Format returns a formatted representation.
func Format(a any) (s string) {
	cfg := spew.ConfigState{
		Indent:                  "    ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
	}
	s = cfg.Sdump(a)
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
