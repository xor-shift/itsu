package util

import "testing"

func TestSliceRemove(t *testing.T) {
	sumInts := func(s []int) int {
		sum := 0
		for _, v := range s {
			sum += v
		}
		return sum
	}

	s := []int{0, 1, 2, 3, 4, 5, 6, 7}

	idx := 5

	sum0 := sumInts(s)
	SliceRemove(&s, idx)
	sum1 := sumInts(s)

	if sum0 != sum1+idx {
		t.Error("sum0 != sum1+idx: ", s)
	}
}

func TestSliceReverse(t *testing.T) {
	s := []int{0, 1, 2, 3, 4}
	SliceReverse(s)

	expected := []int{4, 3, 2, 1, 0}

	for k, v := range s {
		if v != expected[k] {
			t.Error("mismatch at index ", k, ", ", v, " != ", expected[k], " (", s, ")")
			return
		}
	}
}
