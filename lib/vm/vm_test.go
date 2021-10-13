package vm

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestMakeIndex(t *testing.T) {
	vals := []uint64{0, 1, 2, 256, 65534, 65535, 65536, uint64(1<<32) - 1, uint64(1 << 32)}
	expectedTypes := []byte{0, 0, 0, 1, 1, 1, 2, 2, 3}

	for k, v := range vals {
		_, typ := makeIndex(v)

		if typ != expectedTypes[k] {
			t.Errorf("test %d failed: %d", k, typ)
		}
	}
}

func runVM(vm *VM) {
	for {
		vm.DumpNow()
		if err := vm.SingleStep(); err != nil {
			log.Println(err)
			break
		}
	}
}

func TestTokenizeString(t *testing.T) {
	strs := []string{
		"test 123 \"asdasdsad asdasd\" asd\"asd asd",
	}

	tokens := [][]string{
		{"test", "123", "\"asdasdsad asdasd\"", "asd\"asd", "asd"},
	}

	for k, v := range strs {
		r := tokenizeString(v)

		if len(tokens[k]) == len(r) {
			for k2, v2 := range r {
				if v2 != tokens[k][k2] {
					t.Error("test", k, "failed at index", k2, ":", r)
				}
			}
		} else {
			t.Error("test", k, "failed due to slice size (", len(r), "!=", len(tokens[k]), "): ", r)
		}
	}
}

func getDefaultProgram() (b BuiltProgram, err error) {
	builder := NewProgramBuilder()

	if err = CompileFORTH(builder, `
CNUM_const0 1 CMP >=
CNUM_const0 3 CMP <=
AND
"asdasdasd" CSTR_const1 CMP ==
OR
HLT
`); err != nil {
		return
	}

	b = builder.Build()
	return
}

func TestCompile(t *testing.T) {
	var err error
	var built BuiltProgram

	if built, err = getDefaultProgram(); err != nil {
		t.Error(err)
		return
	}

	if linked, err := built.Link(map[string]interface{}{
		"const0": 1.5,
		"const1": "asdasd asd",
	}); err != nil {
		t.Error(err)
		return
	} else {
		vm := NewVM(linked)

		runVM(vm)
	}
}

type valueSerializationPair struct {
	v Value
	b []byte
}

var valueDeserializationPairs = []valueSerializationPair{
	{MakeValue(false), []byte{2, 0, 0}},
	{MakeValue(true), []byte{2, 0, 1}},
	{MakeValue(0), []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	{MakeValue(1), []byte{1, 0, 0, 0, 0, 0, 0, 0, 240, 63}},
	{MakeValue("AAAAAAAA"), []byte{3, 0, 8, 0, 0, 0, 'A', 'A', 'A', 'A', 'A', 'A', 'A', 'A'}},
	{MakeValue(nil), []byte{0, 0}},
}

func TestValue_Serialize(t *testing.T) {
	for k, v := range valueDeserializationPairs {
		b := v.v.Serialize()
		if !reflect.DeepEqual(b, v.b) {
			t.Error(fmt.Sprint("test ", k, " failed: ", b))
		}
	}
}

func TestDeserializeValue(t *testing.T) {
	for k, v := range valueDeserializationPairs {
		reader := bufio.NewReader(bytes.NewReader(v.b))
		dv, err := DeserializeValue(reader)
		if err != nil {
			t.Error("test ", k, " failed: ", err)
		}

		if dv.Kind != v.v.Kind || !reflect.DeepEqual(dv.Data, v.v.Data) {
			t.Error(fmt.Sprint("test ", k, " failed: ", dv))
		}
	}
}

func TestBuiltProgram_Serialize(t *testing.T) {
	var err error
	var built BuiltProgram

	if built, err = getDefaultProgram(); err != nil {
		t.Error(err)
		return
	}

	var built2 BuiltProgram
	var serialized []byte

	if serialized, err = built.Serialize(); err != nil {
		return
	}

	if built2, err = DeserializeBuiltProgram(bufio.NewReader(bytes.NewReader(serialized))); err != nil {
		return
	}

	if !reflect.DeepEqual(built, built2) {
		t.Error("")
	}
}
