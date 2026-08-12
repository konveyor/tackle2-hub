package scm

// Option mask.
type Option byte

// Has returns true when the option (mask) has the option bit set.
func (b Option) Has(opt Option) (has bool) {
	has = b&opt == opt
	return
}

const (
	// CREATE ref on demand enabled.
	CREATE = Option(0x01)
)

// Options builds an option (mask).
func Options(options []Option) (b Option) {
	for _, opt := range options {
		b = b | opt
	}
	return
}

// HasOption returns true when the specified option is enabled.
func HasOption(options []Option, option Option) (has bool) {
	has = Options(options).Has(option)
	return
}
