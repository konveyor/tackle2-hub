package adapter

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/task"
)

var (
	EnvRegex = regexp.MustCompile(`(\$\()([^)]+)(\))`)
)

// EnvInjector inject key into extension metadata.
type EnvInjector struct {
	env  map[string]string
	dict map[string]string
}

// Inject inject into extension metadata.
func (r *EnvInjector) Inject(extension *api.Extension) {
	r.buildEnv(extension)
	mp := make(map[string]any)
	b, _ := json.Marshal(extension.Metadata)
	_ = json.Unmarshal(b, &mp)
	mp = r.inject(mp).(map[string]any)
	extension.Metadata = mp
}

// buildEnv builds the extension `env`.
func (r *EnvInjector) buildEnv(extension *api.Extension) {
	r.env = make(map[string]string)
	for _, env := range extension.Container.Env {
		key := task.ExtEnv(extension.Name, env.Name)
		r.env[env.Name] = os.Getenv(key)
	}
}

// inject replaces both `dict` keys and `env` environment
// variables referenced in metadata.
func (r *EnvInjector) inject(in any) (out any) {
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
			match := EnvRegex.FindStringSubmatch(node)
			if len(match) < 3 {
				break
			}
			k := match[2]
			v := r.env[k]
			if v == "" {
				break
			}
			node = strings.Replace(
				node,
				match[0],
				v,
				-1)
		}
		out = node
	default:
		out = node
	}
	return
}
