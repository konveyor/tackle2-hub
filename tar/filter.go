package tar

import (
	pathlib "path"
	"path/filepath"
)

//
// Filter supports glob-style filtering.
type Filter struct {
	Root    string
	Pattern string
	cache   map[string]bool
}

//
// Match determines if path matches the filter.
func (r *Filter) Match(path string) (b bool) {
	if r.Pattern == "" {
		b = true
		return
	}
	if r.cache == nil {
		r.cache = map[string]bool{}
		matches, _ := filepath.Glob(pathlib.Join(r.Root, r.Pattern))
		for _, p := range matches {
			r.cache[p] = true
		}
	}
	_, b = r.cache[path]
	return
}
