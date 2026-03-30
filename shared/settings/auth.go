package settings

import (
	"github.com/konveyor/tackle2-hub/shared/env"
)

// Environment variables
const (
	EnvAuthRequired = "AUTH_REQUIRED"
)

type Auth struct {
	// Auth required
	Required bool
	// Token settings for builtin provider.
	Token struct {
		Key string
	}
}

func (r *Auth) Load() (err error) {
	r.Required = env.GetBool(EnvAuthRequired, false)
	if !r.Required {
		return
	}
	return
}
