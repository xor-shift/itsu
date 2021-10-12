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

//not an actual test, lol
func TestVM_SingleStep(t *testing.T) {
	builder := NewProgramBuilder()

	idx0 := builder.AddConstantNumber(1.5)
	idx1 := builder.AddConstantString("asdasd asd")

	builder.EmitNLOAD(idx0)
	builder.EmitNCONST(1)
	builder.EmitByte(OpNCMP)
	builder.EmitByte(OpGE)
	builder.EmitNLOAD(idx0)
	builder.EmitNCONST(3)
	builder.EmitByte(OpNCMP)
	builder.EmitByte(OpGE)
	builder.EmitByte(OpLAND)

	builder.EmitSTRLOAD(builder.AddConstantString("asdasdasd"))
	builder.EmitSTRLOAD(idx1)
	builder.EmitByte(OpSCMP)
	builder.EmitByte(OpEQ)

	builder.EmitByte(OpLOR)

	builder.EmitByte(OpHLT)

	vm := NewVM()
	vm.LoadFromBuilder(builder)

	runVM(vm)
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

	constants := map[string]interface{}{
		"const0": 1.5,
		"const1": "asdasd asd",
	}

	if err := CompileFORTH(builder, ""+
		"CNUM_const0 1 CMP >= "+
		"CNUM_const0 3 CMP <= AND "+
		"\"asdasdasd\" CSTR_const1 CMP == OR "+
		"HLT"); err != nil {
		t.Error(err)
		return
	}

	if err := builder.MapConstants(constants); err != nil {
		t.Error(err)
		return
	}

	vm := NewVM()
	vm.LoadFromBuilder(builder)

	runVM(vm)
}
