package profile

import (
	"errors"
	"fmt"
	"strings"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
)

var Log = logr.WithName("task-selector")

type UnknownSelector struct {
	Kind string
}

func (e *UnknownSelector) Error() (s string) {
	return fmt.Sprintf("Selector: '%s' unknown", e.Kind)
}

func (e *UnknownSelector) Is(err error) (matched bool) {
	_, matched = err.(*UnknownSelector)
	return
}

func NewSelector(p crd.ProfileSelector, r Resolver) (selector Selector, err error) {
	match := p.Match
	parsed := ParsedSelector{}
	part := strings.SplitN(match, "/", 2)
	if len(part) > 1 {
		parsed.ns = part[0]
		match = part[1]
	}
	part = strings.SplitN(match, ":", 2)
	if len(part) > 1 {
		parsed.kind = part[0]
		match = part[1]
	}
	part = strings.SplitN(match, "=", 2)
	parsed.name = part[0]
	if len(part) > 1 {
		parsed.value = part[1]
	}
	switch parsed.kind {
	case "":
		selector = &BaseSelector{
			ProfileSelector: p,
			resolver:        r,
			parsed:          parsed,
		}
	case "tag":
		selector = &TagSelector{
			BaseSelector: BaseSelector{
				ProfileSelector: p,
				resolver:        r,
				parsed:          parsed,
			}}
	default:
		err = &UnknownSelector{Kind: parsed.kind}
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

type ParsedSelector struct {
	ns    string
	kind  string
	name  string
	value string
}

type TagSelector struct {
	BaseSelector
}

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
