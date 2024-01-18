package assessment

// NewSet builds a new Set.
func NewSet() (s Set) {
	s = Set{}
	s.members = make(map[uint]bool)
	return
}

// Set is an unordered collection of uints
// with no duplicate elements.
type Set struct {
	members map[uint]bool
}

// Size returns the number of members in the set.
func (r Set) Size() int {
	return len(r.members)
}

// Add a member to the set.
func (r Set) Add(members ...uint) {
	for _, member := range members {
		r.members[member] = true
	}
}

// Contains returns whether an element is a member of the set.
func (r Set) Contains(element uint) bool {
	return r.members[element]
}

// Superset tests whether every element of other is in the set.
func (r Set) Superset(other Set, strict bool) bool {
	if strict && r.Size() <= other.Size() {
		return false
	}
	for m := range other.members {
		if !r.Contains(m) {
			return false
		}
	}
	return true
}

// Subset tests whether every element of this set is in the other.
func (r Set) Subset(other Set, strict bool) bool {
	return other.Superset(r, strict)
}

// Intersects tests whether this set and the other have at least one element in common.
func (r Set) Intersects(other Set) bool {
	for m := range r.members {
		if other.Contains(m) {
			return true
		}
	}
	return false
}

// Members returns the members of the set as a slice.
func (r Set) Members() []uint {
	members := []uint{}
	for k := range r.members {
		members = append(members, k)
	}
	return members
}
