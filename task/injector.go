package task

import (
	"regexp"
	"strconv"
	"strings"

	core "k8s.io/api/core/v1"
)

var (
	SeqRegex = regexp.MustCompile(`(\${seq:)([0-9]+)}`)
)

// Injector macro processor.
type Injector struct {
	seq SeqInjector
}

// Inject process macros.
func (r *Injector) Inject(container *core.Container) {
	var injected []string
	for i := range container.Command {
		if i > 0 {
			injected = append(
				injected,
				r.seq.inject(container.Command[i]))
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
			r.seq.inject(container.Args[i]))
	}
	container.Args = injected
	for i := range container.Env {
		env := &container.Env[i]
		env.Value = r.seq.inject(env.Value)
	}
}

// SeqInjector provides ${seq:<pool>} sequence injection.
type SeqInjector struct {
	portMap map[int]int
}

// inject next integer.
func (r *SeqInjector) inject(in string) (out string) {
	if r.portMap == nil {
		r.portMap = make(map[int]int)
	}
	out = in
	for {
		match := SeqRegex.FindStringSubmatch(out)
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
