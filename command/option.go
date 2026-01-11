package command

import "fmt"

// Options command parameters.
type Options []string

// Add option.
func (a *Options) Add(option string, s ...string) {
	*a = append(*a, option)
	*a = append(*a, s...)
}

// Addf option.
func (a *Options) Addf(option string, x ...interface{}) {
	*a = append(*a, fmt.Sprintf(option, x...))
}
