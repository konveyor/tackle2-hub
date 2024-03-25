package addon

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/konveyor/tackle2-hub/api"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
)

var (
	EnvRegex  = regexp.MustCompile(`(\${)([^}]+)(})`)
	DictRegex = regexp.MustCompile(`(\$\()([^)]+)(\))`)
)

// UnknownInjector used to report an unknown injector.
type UnknownInjector struct {
	Kind string
}

func (e *UnknownInjector) Error() (s string) {
	return fmt.Sprintf("Injector: %s, unknown.", e.Kind)
}

func (e *UnknownInjector) Is(err error) (matched bool) {
	var inst *UnknownInjector
	matched = errors.As(err, &inst)
	return
}

// UnknownField used to report an unknown resource field..
type UnknownField struct {
	Kind  string
	Field string
}

func (e *UnknownField) Error() (s string) {
	return fmt.Sprintf("Field: %s.%s, unknown.", e.Kind, e.Field)
}

func (e *UnknownField) Is(err error) (matched bool) {
	var inst *UnknownField
	matched = errors.As(err, &inst)
	return
}

// ResourceInjector supports hub resource injection.
type ResourceInjector struct {
	dict map[string]string
}

// Build returns  the built resource dictionary.
func (r *ResourceInjector) Build(task *Task, extension *api.Extension) (dict map[string]string, err error) {
	richClient := task.richClient
	r.dict = make(map[string]string)
	for _, injector := range extension.Resources {
		parsed := strings.Split(injector.Kind, "=")
		switch strings.ToLower(parsed[0]) {
		case "identity":
			kind := ""
			if len(parsed) > 1 {
				kind = parsed[1]
			}
			id := task.task.Application.ID
			identity, found, nErr := richClient.Application.FindIdentity(id, kind)
			if nErr != nil {
				err = nErr
				return
			}
			if found {
				err = r.add(&injector, identity)
				if err != nil {
					return
				}
			}
		default:
			err = &UnknownInjector{Kind: parsed[0]}
			return
		}
	}
	dict = r.dict
	return
}

// add the resource fields specified in the injector.
func (r *ResourceInjector) add(injector *crd.Injector, object any) (err error) {
	objectMap := r.objectMap(object)
	for _, f := range injector.Fields {
		v, found := objectMap[f.Name]
		if !found {
			err = &UnknownField{Kind: injector.Kind, Field: f.Name}
			return
		}
		fv := r.string(v)
		if f.Path != "" {
			err = r.write(f.Path, fv)
			if err != nil {
				return
			}
			fv = f.Path
		}
		r.dict[f.Key] = fv
	}
	return
}

// write a resource field value to a file.
func (r *ResourceInjector) write(path string, s string) (err error) {
	f, err := os.Create(path)
	if err == nil {
		_, _ = f.Write([]byte(s))
		_ = f.Close()
	}
	return
}

// string returns a string representation of a field value.
func (r *ResourceInjector) string(object any) (s string) {
	if object != nil {
		s = fmt.Sprintf("%v", object)
	}
	return
}

// objectMap returns a map for a resource object.
func (r *ResourceInjector) objectMap(object any) (mp map[string]any) {
	b, _ := json.Marshal(object)
	mp = make(map[string]any)
	_ = json.Unmarshal(b, &mp)
	return
}

// MetaInjector inject key into extension metadata.
type MetaInjector struct {
	env  map[string]string
	dict map[string]string
}

// Inject inject into extension metadata.
func (r *MetaInjector) Inject(extension *api.Extension) {
	r.buildEnv(extension)
	mp := make(map[string]any)
	b, _ := json.Marshal(extension.Metadata)
	_ = json.Unmarshal(b, &mp)
	mp = r.inject(mp).(map[string]any)
	extension.Metadata = mp
}

// buildEnv builds the `env`.
// Maps EXTENSION_<extension>_<envar> found in the addon environment to its
// original unqualified name in the extension environment.
func (r *MetaInjector) buildEnv(extension *api.Extension) {
	r.env = make(map[string]string)
	for _, env := range extension.Container.Env {
		key := strings.Join(
			[]string{
				"EXTENSION",
				strings.ToUpper(extension.Name),
				env.Name,
			},
			"_")
		r.env[env.Name] = os.Getenv(key)
	}
}

// inject replaces both `dict` keys and `env` environment
// variables referenced in metadata.
func (r *MetaInjector) inject(in any) (out any) {
	switch node := in.(type) {
	case map[string]any:
		for k, v := range node {
			node[k] = r.inject(v)
		}
		out = node
	case []any:
		var injected []any
		for _, n := range node {
			injected = append(
				injected,
				r.inject(n))
		}
		out = injected
	case string:
		for {
			match := DictRegex.FindStringSubmatch(node)
			if len(match) < 3 {
				break
			}
			node = strings.Replace(
				node,
				match[0],
				r.dict[match[2]],
				-1)
		}
		for {
			match := EnvRegex.FindStringSubmatch(node)
			if len(match) < 3 {
				break
			}
			node = strings.Replace(
				node,
				match[0],
				r.env[match[2]],
				-1)
		}
		out = node
	default:
		out = node
	}
	return
}
