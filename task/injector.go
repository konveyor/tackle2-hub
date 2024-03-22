package task

import (
	"regexp"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"
)

var (
	PortRegex = regexp.MustCompile(`(\$\(port:)([0-9]+)\)`)
)

// Injector macro processor.
type Injector struct {
	port PortInjector
}

// Inject process macros.
func (r *Injector) Inject(container *core.Container) {
	var injected []string
	for i := range container.Command {
		if i > 0 {
			injected = append(
				injected,
				r.port.inject(container.Command[i]))
		} else {
			injected = append(
				injected,
				container.Command[i])
		}
	}
	container.Command = injected
	injected = nil
	for i := range container.Args {
		injected = append(
			injected,
			r.port.inject(container.Args[i]))
	}
	container.Args = injected
	for i := range container.Env {
		env := &container.Env[i]
		env.Value = r.port.inject(env.Value)
	}
}

// PortInjector provides $(port:<base>) injection.
type PortInjector struct {
	portMap map[int]int
}

// inject port.
func (r *PortInjector) inject(in string) (out string) {
	if r.portMap == nil {
		r.portMap = make(map[int]int)
	}
	out = in
	for {
		match := PortRegex.FindStringSubmatch(out)
		if len(match) < 3 {
			break
		}
		base, _ := strconv.Atoi(match[2])
		offset := r.portMap[base]
		out = strings.Replace(
			out,
			match[0],
			strconv.Itoa(base+offset),
			-1)
		offset++
		r.portMap[base] = offset
	}
	return
}
