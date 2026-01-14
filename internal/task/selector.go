package task

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	gv "github.com/PaesslerAG/gval"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/ptr"
	"gorm.io/gorm"
)

var (
	PredRegex = regexp.MustCompile(`(\w+):(\S+)`)
)

// NewSelector returns a selector.
func NewSelector(db *gorm.DB, task *Task) (selector *Selector) {
	selector = &Selector{
		predicate: map[string]Predicate{
			"tag": &TagPredicate{
				db:   db,
				task: task,
			},
			"platform": &PlatformPredicate{
				db:   db,
				task: task,
			},
		},
	}
	return
}

// Selector used to match addons and extensions.
type Selector struct {
	predicate map[string]Predicate
}

// Match evaluates the selector.
func (r *Selector) Match(selector string) (matched bool, err error) {
	if selector == "" {
		matched = true
		return
	}
	params := make(map[string]string)
	found := PredRegex.FindAllStringSubmatch(selector, -1)
	for _, m := range found {
		kind := m[1]
		p, found := r.predicate[kind]
		if found {
			matched, err = p.Match(m[2])
			if err != nil {
				return
			}
			params[m[0]] = strconv.FormatBool(matched)
		} else {
			err = &SelectorNotValid{
				Selector:  selector,
				Predicate: kind,
			}
			return
		}
	}
	var keySet []string
	for k := range params {
		keySet = append(keySet, k)
	}
	sort.Slice(
		keySet,
		func(i, j int) bool {
			return len(keySet[i]) > len(keySet[j])
		})
	matched = false
	expression := selector
	for _, ref := range keySet {
		expression = strings.Replace(
			expression,
			ref,
			params[ref],
			-1)
	}
	p := r.parser()
	v, err := p.Evaluate(expression, nil)
	if err != nil {
		err = &SelectorNotValid{
			Selector: selector,
			Reason:   err.Error(),
		}
		return
	}
	if b, cast := v.(bool); cast {
		matched = b
	} else {
		err = &SelectorNotValid{
			Selector: selector,
			Reason:   "parser returned unexpected result.",
		}
	}
	return
}

// parser returns a parser.
func (r *Selector) parser() (p gv.Language) {
	p = gv.NewLanguage(
		gv.Ident(),
		gv.Parentheses(),
		gv.Constant("true", true),
		gv.Constant("false", false),
		gv.PrefixOperator(
			"!",
			func(c context.Context, v any) (b any, err error) {
				switch x := v.(type) {
				case bool:
					b = !x
				default:
					err = &SelectorNotValid{
						Reason: fmt.Sprintf("%v not expected", x),
					}
				}
				return
			}),
		gv.InfixShortCircuit(
			"&&",
			func(a any) (v any, b bool) {
				v = false
				b = a == false
				return
			}),
		gv.InfixBoolOperator(
			"&&",
			func(a, b bool) (v any, err error) {
				v = a && b
				return
			}),
		gv.InfixShortCircuit(
			"||",
			func(a any) (v any, b bool) {
				v = true
				b = a == true
				return
			}),
		gv.InfixBoolOperator(
			"||",
			func(a, b bool) (v any, err error) {
				v = a || b
				return
			}),
	)
	return
}

type Predicate interface {
	Match(ref string) (matched bool, err error)
}

// TagPredicate evaluates application tag references.
type TagPredicate struct {
	db   *gorm.DB
	task *Task
}

// Match evaluates application tag references.
// The `ref` has format: category=tag.
// The tag and behaves like a wildcard when not specified.
func (r *TagPredicate) Match(ref string) (matched bool, err error) {
	category, name := r.parse(ref)
	db := r.db.Session(&gorm.Session{})
	cat := &model.TagCategory{}
	err = db.First(cat, "name=?", category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Log.Info(
				"TagSelector: category not found.",
				"name",
				category)
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	m := &model.Application{}
	appId := ptr.ID(r.task.ApplicationID)
	db = r.db.Session(&gorm.Session{})
	db = db.Preload("Tags")
	err = db.First(m, appId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Log.Info(
				"TagSelector: application not found.",
				"id",
				appId)
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	for _, tag := range m.Tags {
		if cat.ID != tag.CategoryID {
			continue
		}
		if !(name == "" || tag.Name == name) {
			continue
		}
		matched = true
		break
	}
	return
}

// parse tag ref.
func (r *TagPredicate) parse(s string) (category, name string) {
	part := strings.SplitN(s, "=", 2)
	category = part[0]
	if len(part) > 1 {
		name = part[1]
	}
	return
}

// PlatformPredicate evaluates application tag references.
type PlatformPredicate struct {
	db   *gorm.DB
	task *Task
}

// Match evaluates application tag references.
// The `ref` has format: kind=<kind>.
func (r *PlatformPredicate) Match(ref string) (matched bool, err error) {
	key, value := r.parse(ref)
	switch key {
	case "kind":
		// supported.
	default:
		Log.Info(
			"PlatformSelector: key not supported.",
			"key",
			key)
		return
	}
	m := &model.Platform{}
	platformId := ptr.ID(r.task.PlatformID)
	db := r.db.Session(&gorm.Session{})
	db = db.Where("id", platformId)
	db = db.Where(key, value)
	err = db.First(m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		} else {
			err = liberr.Wrap(err)
		}
		return
	}
	matched = true
	return
}

// parse kind ref.
func (r *PlatformPredicate) parse(s string) (key, kind string) {
	part := strings.SplitN(s, "=", 2)
	key = part[0]
	if len(part) > 1 {
		kind = part[1]
	}
	return
}
