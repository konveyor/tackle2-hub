package profile

import (
	"strings"

	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

func NewSelector(p crd.ProfileSelector, r Resolver) (selector Selector, err error) {
	match := p.Match
	parsed := ParsedSelector{}
	part := strings.SplitN(match, "/", 1)
	if len(part) > 1 {
		parsed.ns = part[0]
		match = part[1]
	}
	part = strings.SplitN(match, ":", 1)
	if len(part) > 1 {
		parsed.kind = part[0]
		match = part[1]
	}
	part = strings.SplitN(match, "=", 1)
	parsed.name = part[0]
	if len(part) > 1 {
		parsed.value = part[1]
	}
	switch parsed.kind {
	case "tag":
		selector = &TagSelector{
			BaseSelector: BaseSelector{
				ProfileSelector: p,
				resolver:        r,
				parsed:          parsed,
			}}
	default:
	}
	return
}

type Selector interface {
	Match(db *gorm.DB, task *model.Task) (matched []string, err error)
}

type BaseSelector struct {
	crd.ProfileSelector
	resolver Resolver
	parsed   ParsedSelector
}

type ParsedSelector struct {
	ns    string
	kind  string
	name  string
	value string
}

type TagSelector struct {
	BaseSelector
}

func (r *TagSelector) Match(db *gorm.DB, task *model.Task) (matched []string, err error) {
	parsed := r.parsed
	db = db.Session(&gorm.Session{})
	application := &model.Application{}
	err = db.First(application, task.Application.ID).Error
	if err != nil {
		return
	}
	db = db.Session(&gorm.Session{})
	category := &model.TagCategory{}
	err = db.First(category, "name=?", parsed.name).Error
	if err != nil {
		return
	}
	for _, ref := range application.Tags {
		tag := &model.Tag{}
		err = db.First(tag, ref.ID).Error
		if err != nil {
			return
		}
		if parsed.name == "" || tag.Name == parsed.name {
			if r.Name != "" {
				matched = append(matched, r.Name)
			}
			if r.Capability != "" {
				var names []string
				names, err = r.resolver.Match(r.Capability)
				if err == nil {
					matched = append(matched, names...)
				} else {
					return
				}
			}

		}
	}
	return
}
