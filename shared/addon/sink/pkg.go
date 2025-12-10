package sink

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/konveyor/tackle2-hub/shared/addon/adapter"
)

var (
	addon = adapter.Addon
)

func New(enabled bool) logr.LogSink {
	return &Sink{enabled: enabled}
}

// Sink used to bridge a Logger to the addon.Activity.
type Sink struct {
	enabled bool
}

func (s *Sink) Init(_ logr.RuntimeInfo) {
}

func (s *Sink) Enabled(_ int) (enabled bool) {
	enabled = s.enabled
	return
}

func (s *Sink) Info(_ int, msg string, kv ...any) {
	msg = s.join(msg, kv)
	addon.Activity(msg)
	return
}

func (s *Sink) Error(err error, msg string, kv ...any) {
	msg = s.join(msg, kv)
	msg += "\n"
	msg += err.Error()
	addon.Activity(msg)
	return
}

func (s *Sink) WithValues(_ ...any) logr.LogSink {
	return s
}

func (s *Sink) WithName(_ string) logr.LogSink {
	return s
}

func (s *Sink) join(m string, kv ...any) (joined string) {
	items := []string{}
	for i := range kv {
		if i%2 != 0 {
			key := fmt.Sprintf("%v", kv[i-1])
			v := fmt.Sprintf("%+v", kv[i])
			p := fmt.Sprintf("%s=%s", key, v)
			items = append(items, p)
		}
	}
	joined = m
	if len(items) > 0 {
		joined += ":"
		joined += strings.Join(items, ",")
	}
	return
}
