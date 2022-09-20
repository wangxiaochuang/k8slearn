package slices

func Equal(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, n := range s1 {
		if n != s2[i] {
			return false
		}
	}
	return true
}

func Filter(d, s []string, keep func(string) bool) []string {
	for _, n := range s {
		if keep(n) {
			d = append(d, n)
		}
	}
	return d
}

func Contains(s []string, v string) bool {
	return Index(s, v) >= 0
}

func Index(s []string, v string) int {
	// "Contains" may be replaced with "Index(s, v) >= 0":
	// https://github.com/golang/go/issues/45955#issuecomment-873377947
	for i, n := range s {
		if n == v {
			return i
		}
	}
	return -1
}

func Clone(s []string) []string {
	// https://github.com/go101/go101/wiki/There-is-not-a-perfect-way-to-clone-slices-in-Go
	if s == nil {
		return nil
	}
	c := make([]string, len(s))
	copy(c, s)
	return c
}
