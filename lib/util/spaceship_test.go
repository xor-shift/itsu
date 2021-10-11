package util

import "testing"

func TestSpaceship(t *testing.T) {
	s0s := []interface{}{0, 1, 0, "A", "B", "A", 0, 0}
	s1s := []interface{}{0, 0, 1, "A", "A", "B", uint(0), "A"}
	res := []int{0, 1, -1, 0, 1, -1, 2, 2}

	for k, e := range res {
		r := Spaceship(s0s[k], s1s[k])
		if e != r {
			t.Error("test ", k, " failed: ", r)
		}
	}
}
