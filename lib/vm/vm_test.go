package vm

import (
	"log"
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

func TestCompile(t *testing.T) {
	builder := NewProgramBuilder()

	if err := CompileFORTH(builder, `
CNUM_const0 1 CMP >=
CNUM_const0 3 CMP <=
AND
"asdasdasd" CSTR_const1 CMP ==
OR
HLT
`); err != nil {
		t.Error(err)
		return
	}

	built := builder.Build()
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