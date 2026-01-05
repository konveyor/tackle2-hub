package ptr

// ID safely returns the id.
func ID(p *uint) (id uint) {
	if p != nil {
		id = *p
	}
	return
}
