package util

func Strcmp(s0, s1 string) rune {
	i0 := 0
	i1 := 0

	r0 := append([]rune(s0), 0)
	r1 := append([]rune(s1), 0)

	for r0[i0] != 0 && (r0[i0] == r1[i1]) {
		i0++
		i1++
	}

	return r0[i0] - r1[i1]
}
