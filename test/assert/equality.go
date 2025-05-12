package assert

import (
	"fmt"
	"sort"
)

type Map = map[string]any

// Simple equality check working for flat types (no nested types passed by reference).
func FlatEqual(got, expected any) bool {
	return fmt.Sprintf("%v", got) == fmt.Sprintf("%v", expected)
}

// MapEq compares maps.
func MapEq(a, b Map) (eq bool) {
	defer func() {
		_ = recover()
	}()
	if len(a) != len(b) {
		return
	}
	var keyset []string
	for k := range a {
		keyset = append(keyset, k)
	}
	sort.Strings(keyset)
	for _, k := range keyset {
		vA := a[k]
		vB := b[k]
		switch vA.(type) {
		case Map:
			eq = MapEq(vA.(Map), vB.(Map))
			if !eq {
				return
			}
		default:
			eq = FlatEqual(vA, vB)
			if !eq {
				return
			}
		}
		if a[k] != b[k] {
			return
		}
	}
	eq = true
	return
}
