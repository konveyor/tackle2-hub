package settings

import (
	"fmt"
	"os"
	"strconv"

	"github.com/konveyor/tackle2-hub/shared/env"
)

// Environment variables.
const (
	MetricsEnabled = "METRICS_ENABLED"
	MetricsPort    = "METRICS_PORT"
)

// Metrics settings
type Metrics struct {
	// Metrics port.
	Port int
	// Metrics enabled.
	Enabled bool
}

// Load settings.
func (r *Metrics) Load() error {
	// Enabled
	r.Enabled = env.GetBool(MetricsEnabled, true)
	// Port
	if s, found := os.LookupEnv(MetricsPort); found {
		r.Port, _ = strconv.Atoi(s)
	} else {
		r.Port = 2112
	}

	return nil
}

// Address on which to serve metrics.
func (r *Metrics) Address() string {
	return fmt.Sprintf(":%d", r.Port)
}
