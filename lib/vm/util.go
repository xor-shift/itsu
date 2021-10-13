package vm

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"strings"
	"unicode"
)

func makeIndex(index uint64) ([]byte, uint8) {
	var v interface{}
	var iType uint8

	if index < uint64(1<<8) {
		v = uint8(index)
		iType = 0
	} else if index < uint64(1<<16) {
		v = uint16(index)
		iType = 1
	} else if index < uint64(1<<32) {
		v = uint32(index)
		iType = 2
	} else {
		v = index
		iType = 3
	}

	buf := make([]byte, binary.Size(v))
	_ = binary.Write(bytes.NewBuffer(buf), binary.BigEndian, v)

	return buf, iType
}

func tokenizeString(str string) []string {
	const (
		stateInit          = 0
		stateReadingAtom   = 1
		stateReadingString = 2
	)

	state := stateInit
	inEscape := false
	buffer := strings.Builder{}
	tokens := make([]string, 0)

	pushRune := func(r rune) {
		buffer.WriteRune(r)
	}

	endToken := func() {
		state = stateInit

		if buffer.Len() == 0 {
			return
		}

		tokens = append(tokens, buffer.String())
		buffer = strings.Builder{}
	}

	s := map[int]func(r rune){
		stateInit: func(r rune) {
			if unicode.IsSpace(r) {
				return
			} else {
				pushRune(r)

				if r == '"' {
					state = stateReadingString
				} else {
					state = stateReadingAtom
				}
			}
		},
		stateReadingAtom: func(r rune) {
			if inEscape {
				pushRune(r)
				inEscape = false
			} else {
				if unicode.IsSpace(r) {
					endToken()
				} else if r == '\\' {
					inEscape = true
				} else {
					pushRune(r)
				}
			}
		},
		stateReadingString: func(r rune) {
			if inEscape {
				pushRune(r)
				inEscape = false
			} else {
				if r == '"' {
					pushRune(r)
					endToken()
				} else if r == '\\' {
					inEscape = true
				} else {
					pushRune(r)
				}
			}
		},
	}

	for _, r := range []rune(str) {
		s[state](r)
	}
	endToken()

	return tokens
}

func genericPush(slicePtr interface{}, ptr *int, v interface{}) error {
	sl := reflect.ValueOf(slicePtr)
	slv := reflect.Indirect(sl)

	if *ptr == slv.Len() {
		return ErrorOverflow
	}

	slv.Index(*ptr).Set(reflect.ValueOf(v))
	*ptr += 1

	return nil
}

func genericTop(slice interface{}, sp int) (interface{}, error) {
	if sp == 0 {
		return nil, ErrorUnderflow
	}

	return reflect.ValueOf(slice).Index(sp - 1), nil
}

func genericPop(slicePtr interface{}, ptr *int) (interface{}, error) {
	if *ptr == 0 {
		return nil, ErrorUnderflow
	}

	sl := reflect.ValueOf(slicePtr)
	slv := reflect.Indirect(sl)

	*ptr -= 1
	val := slv.Index(*ptr).Interface()

	return val, nil
}
