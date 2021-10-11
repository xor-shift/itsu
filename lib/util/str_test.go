package util

import "testing"

func TestStrcmp(t *testing.T) {
	s0s := []string{"AAAA", "AAAB", "AAAA", "AAA", "AAAC"}
	s1s := []string{"AAAA", "AAAA", "AAAC", "AAAB", "ACC"}

	res := []int{0, 1, -1, -1, -1}

	for k, e := range res {
		r := Strcmp(s0s[k], s1s[k])
		pass := (r < 0 && e < 0) ||
			(r > 0 && e > 0) ||
			(r == 0 && e == 0)

		if !pass {
			t.Error("test ", k, "failed: ", r)
		}
	}
}
