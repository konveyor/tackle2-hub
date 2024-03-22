package task

import (
	"regexp"
	"strconv"
	"strings"

	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
)

// Injector macro processor.
type Injector struct {
	port PortInjector
}

// Inject process macros.
func (r *Injector) Inject(extension *crd.Extension) {
	container := &extension.Spec.Container
	for _, env := range container.Env {
		env.Value = r.port.inject(env.Value)
	}
	r.port.next()
}

// Port port allocation.
type Port struct {
	next    int
	matched int
}

// PortInjector provides $(port:<base>) injection.
type PortInjector struct {
	pattern *regexp.Regexp
	portMap map[int]Port
}

// init object.
func (r *PortInjector) init() {
	if r.pattern != nil {
		return
	}
	r.pattern = regexp.MustCompile(`(\$\(port:)([0-9]+)\)`)
	r.portMap = make(map[int]Port)
}

// inject port.
func (r *PortInjector) inject(in string) (out string) {
	r.init()
	out = in
	for {
		match := r.pattern.FindStringSubmatch(out)
		if len(match) < 3 {
			break
		}
		base, _ := strconv.Atoi(match[2])
		port := r.portMap[base]
		out = strings.Replace(
			out,
			match[0],
			strconv.Itoa(base+port.next),
			-1)
		port.matched++
		r.portMap[base] = port
	}
	return
}

// next increment offsets and reset matched.
func (r *PortInjector) next() {
	r.init()
	for base, port := range r.portMap {
		if port.matched > 0 {
			port.matched = 0
			port.next++
			r.portMap[base] = port
		}
	}
}
