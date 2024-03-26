package task

import (
	"errors"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

// NewSelector returns a configured selector.
func NewSelector(p crd.Selector, r Resolver) (selector Selector, err error) {
	parsed := ParsedSelector{}
	parsed.With(p.Match)
	switch parsed.kind {
	case "":
		selector = &BaseSelector{
			Selector: p,
			resolver: r,
			parsed:   parsed,
		}
	case "tag":
		selector = &TagSelector{
			BaseSelector: BaseSelector{
				Selector: p,
				resolver: r,
				parsed:   parsed,
			}}
	default:
		err = &SelectorNotSupported{
			Kind: parsed.kind,
		}
	}
	return
}

// Selector find resources based on criteria.
type Selector interface {
	// Match returns resources matching a criteria.
	Match(db *gorm.DB, task *model.Task) (matched []string, err error)
}

// BaseSelector -
type BaseSelector struct {
	crd.Selector
	resolver Resolver
	parsed   ParsedSelector
}

// Match returns resources directly by name or criteria.
func (r *BaseSelector) Match(db *gorm.DB, task *model.Task) (matched []string, err error) {
	if r.Name != "" {
		if r.resolver.Find(r.Name) {
			matched = []string{r.Name}
		}
		return
	}
	if r.Capability != "" {
		matched, err = r.resolver.Match(r.Capability)
		return
	}
	return
}

// ParsedSelector -
type ParsedSelector struct {
	ns    string
	kind  string
	name  string
	value string
}

// With parses and populates the selector.
func (p *ParsedSelector) With(s string) {
	part := strings.SplitN(s, "/", 2)
	if len(part) > 1 {
		p.ns = part[0]
		s = part[1]
	}
	part = strings.SplitN(s, ":", 2)
	if len(part) > 1 {
		p.kind = part[0]
		s = part[1]
	}
	part = strings.SplitN(s, "=", 2)
	p.name = part[0]
	if len(part) > 1 {
		p.value = part[1]
	}
}

// TagSelector matches resources by tag.
type TagSelector struct {
	BaseSelector
}

// Match returns resources by matching the tags.
// Format: tag:<category>=<tag>
// When <tag> is not specified, $* represents the tag.
// Example:
//   - match: platform:target=
//     capability: $*-analysis
func (r *TagSelector) Match(dbIn *gorm.DB, task *model.Task) (matched []string, err error) {
	parsed := r.parsed
	db := dbIn.Session(&gorm.Session{})
	cat := &model.TagCategory{}
	err = db.First(cat, "name=?", parsed.name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Log.Info(
				"TagSelector: category not found.",
				"name",
				parsed.name)
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	db = dbIn.Session(&gorm.Session{})
	db = db.Preload("Tags")
	application := &model.Application{}
	err = db.First(application, task.ApplicationID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Log.Info(
				"TagSelector: application not found.",
				"id",
				task.ApplicationID)
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	for _, tag := range application.Tags {
		if cat.ID != tag.CategoryID {
			continue
		}
		if !(parsed.value == "" || tag.Name == parsed.value) {
			continue
		}
		if r.Name != "" {
			name := strings.Replace(
				r.Name,
				"$*",
				strings.ToLower(tag.Name),
				1)
			if r.resolver.Find(name) {
				matched = append(matched, name)
			}
		}
		if r.Capability != "" {
			var names []string
			capability := strings.Replace(
				r.Capability,
				"$*",
				tag.Name,
				1)
			names, err = r.resolver.Match(capability)
			if err == nil {
				matched = append(matched, names...)
			} else {
				return
			}
		}
	}
	return
}
