package convert

func Intersect(a, b []int64) []int64 {
	m := make(map[int64]bool)
	n := make([]int64, 0)

	for _, v := range a {
		m[v] = true
	}

	for _, v := range b {
		if m[v] {
			n = append(n, v)
		}
	}

	return n
}

func IntersectN(s ...[]int64) []int64 {
	if len(s) == 0 {
		return nil
	}
	result := s[0]
	for i := 1; i < len(s); i++ {
		result = Intersect(result, s[i])
	}
	return result
}
