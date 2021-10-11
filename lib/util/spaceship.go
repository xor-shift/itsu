package util

import (
	"reflect"
)

//Spaceship compares two interfaces, (<, ==, >) = (-1, 0, 1). 2 means that the values are incomparable
func Spaceship(lhs, rhs interface{}) (res int) {
	defer func() {
		if r := recover(); r != nil {
			res = 2
		}
	}()

	vLhs := reflect.ValueOf(lhs)
	vRhs := reflect.ValueOf(rhs)

	tLhs := vLhs.Type().Kind()
	//tRhs := vRhs.Type().Kind()

	reflect.DeepEqual(lhs, rhs)

	switch tLhs {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v0 := vLhs.Int()
		v1 := vRhs.Int()
		if v0 == v1 {
			res = 0
		} else if v0 > v1 {
			res = 1
		} else {
			res = -1
		}
		break
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v0 := vLhs.Uint()
		v1 := vRhs.Uint()
		if v0 < v1 {
			res = -1
		} else if v0 == v1 {
			res = 0
		} else {
			res = 1
		}
		break
	case reflect.String:
		v0 := vLhs.String()
		v1 := vRhs.String()
		tRes := Strcmp(v0, v1)
		if tRes < 0 {
			res = -1
		} else if tRes == 0 {
			res = 0
		} else {
			res = 1
		}
		break
	default:
		res = 2
		break
	}

	return
}
