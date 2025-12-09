package tar

import (
	pathlib "path"
	"path/filepath"
)

// NewFilter returns a filter.
func NewFilter(root string) (f Filter) {
	f = Filter{Root: root}
	return
}

// Filter supports glob-style filtering.
type Filter struct {
	included FilterSet
	excluded FilterSet
	Root     string
}

// Match determines if path matches the filter.
func (r *Filter) Match(path string) (b bool) {
	r.included.root = r.Root
	r.excluded.root = r.Root
	if r.included.Len() > 0 {
		included := r.included.Match(path)
		if !included {
			return
		}
	}
	b = true
	if r.excluded.Len() > 0 {
		excluded := r.excluded.Match(path)
		if excluded {
			b = false
			return
		}
	}
	return
}

// Include adds included patterns.
// Empty ("") patterns are ignored.
func (r *Filter) Include(patterns ...string) {
	r.included.Add(patterns...)
}

// Exclude adds excluded patterns.
// Empty ("") patterns are ignored.
func (r *Filter) Exclude(patterns ...string) {
	r.excluded.Add(patterns...)
}

// FilterSet is a collection of filter patterns.
type FilterSet struct {
	root     string
	patterns []string
	cache    map[string]bool
}

// Match returns true when the path matches.
func (r *FilterSet) Match(path string) (match bool) {
	r.build()
	_, match = r.cache[path]
	return
}

// Add pattern.
// Empty ("") patterns are ignored.
func (r *FilterSet) Add(patterns ...string) {
	for _, p := range patterns {
		if p == "" {
			continue
		}
		r.cache = nil
		r.patterns = append(
			r.patterns,
			p)
	}
}

// Len returns number of patterns.
func (r *FilterSet) Len() (n int) {
	return len(r.patterns)
}

// build populates the cache as needed.
func (r *FilterSet) build() {
	if r.cache != nil {
		return
	}
	r.cache = make(map[string]bool)
	for i := range r.patterns {
		matches, _ := filepath.Glob(pathlib.Join(r.root, r.patterns[i]))
		for _, p := range matches {
			r.cache[p] = true
		}
	}
}
