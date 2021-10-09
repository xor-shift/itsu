package util

import (
	"reflect"
)

//SliceRemove removes an element from a slice while not preserving the order of the elements, the last element will be brought to the given index.
//sPtr must be a pointer to a slice
func SliceRemove(sPtr interface{}, index int) {
	vp := reflect.ValueOf(sPtr)
	v := vp.Elem()

	v.Index(index).Set(v.Index(v.Len() - 1))
	v.Set(v.Slice(0, v.Len()-1))
}

func SliceReverse(s interface{}) {
	v := reflect.ValueOf(s)
	swapper := reflect.Swapper(s)

	for i := 0; i < v.Len()/2; i++ {
		swapper(i, v.Len()-i-1)
	}
}

/* i wish i had generics
void ArrRemove(auto &v, size_t index) {
	std::swap(v[index], v[v.size() - 1]);
	v.resize(v.size() - 1);
}

func ArrReverse(auto &v) {
	for (size_t i = 0; i < v.size() / 2; i++)
		std::swap(v[i], v[v.size() - i - 1];
}
*/
