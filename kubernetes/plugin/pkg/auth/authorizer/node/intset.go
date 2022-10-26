package node

// intSet maintains a map of id to refcounts
type intSet struct {
	// members is a map of id to refcounts
	members map[int]int
}

func newIntSet() *intSet {
	return &intSet{members: map[int]int{}}
}

// has returns true if the specified id has a positive refcount.
// it is safe to call concurrently, but must not be called concurrently with any of the other methods.
func (s *intSet) has(i int) bool {
	if s == nil {
		return false
	}
	return s.members[i] > 0
}

// reset removes all ids, effectively setting their refcounts to 0.
// it is not thread-safe.
func (s *intSet) reset() {
	for k := range s.members {
		delete(s.members, k)
	}
}

// increment adds one to the refcount of the specified id.
// it is not thread-safe.
func (s *intSet) increment(i int) {
	s.members[i]++
}

// decrement removes one from the refcount of the specified id,
// and removes the id if the resulting refcount is <= 0.
// it will not track refcounts lower than zero.
// it is not thread-safe.
func (s *intSet) decrement(i int) {
	if s.members[i] <= 1 {
		delete(s.members, i)
	} else {
		s.members[i]--
	}
}
