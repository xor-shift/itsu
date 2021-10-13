package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"example.com/itsuMain/lib/util"
	"fmt"
	"math"
)

const (
	stackSize     = 64
	callStackSize = 16
)

var (
	ErrorBadEOF              = errors.New("program ended prematurely")
	ErrorEOF                 = errors.New("program has finished")
	ErrorHLT                 = errors.New("vm is halted")
	ErrorBadOpcode           = errors.New("unhandled opcode")
	ErrorOverflow            = errors.New("stack overflow")
	ErrorUnderflow           = errors.New("stack underflow")
	ErrorUnsupportedConstant = errors.New("loaded constant's type is not supported")
	ErrorType                = errors.New("type error")
	ErrorArithmetic          = errors.New("arithmetic sign error")
)

type callFrame struct {
	rp   int
	vars map[uint32]Value
}

type VM struct {
	program Program
	pc      int

	stack [stackSize]Value
	sp    int

	callStack [callStackSize]callFrame
	csp       int

	halt bool
}

func NewVM(program Program) *VM {
	v := &VM{
		program: program,
		pc:      0,
		stack:   [stackSize]Value{ValueNil},
		sp:      0,
	}

	return v
}

func (vm *VM) SingleStep() (err error) {
	if vm.pc == len(vm.program.Program) {
		return ErrorEOF
	} else if vm.halt {
		return ErrorHLT
	}

	var opcode byte
	var argBytes []byte
	if opcode, argBytes, err = vm.readInstruction(); err != nil {
		return
	}

	argReader := bytes.NewReader(argBytes)

	index := uint32(0)
	number := float64(0)

	if argSize := GetOpcodeProperties(opcode).ArgSize; argSize == 4 {
		if err = binary.Read(argReader, binary.LittleEndian, &index); err != nil {
			return
		}
	} else if argSize == 8 {
		if err = binary.Read(argReader, binary.LittleEndian, &number); err != nil {
			return
		}
	}

	arithHelper := func(fn func(lhs, rhs float64) float64) (err error) {
		if lhs, rhs, err := vm.Pop2Kind(KindNumber); err != nil {
			return err
		} else {
			return vm.Push(MakeValue(fn(lhs.Data.(float64), rhs.Data.(float64))))
		}
	}

	unaryArithHelper := func(fn func(v float64) float64) (err error) {
		if v, err := vm.PopKind(KindNumber); err != nil {
			return err
		} else {
			return vm.Push(MakeValue(fn(v.Data.(float64))))
		}
	}

	opcodeHandlers := map[uint8]func() error{
		OpNCONST:   func() error { return vm.Push(MakeValue(number)) },
		OpNCONST_0: func() error { return vm.Push(MakeValue(opcode - OpNCONST_0)) },
		OpBCONST_0: func() error { return vm.Push(MakeValue(opcode == OpBCONST_1)) },
		OpNILCONST: func() error { return vm.Push(MakeValue(nil)) },
		OpISNIL: func() error {
			if vm.sp == 0 {
				return ErrorUnderflow
			}

			return vm.Push(MakeValue(vm.stack[vm.sp-1].Kind == KindNil))
		},
		OpKIND: func() error {
			if vm.sp == 0 {
				return ErrorUnderflow
			}

			return vm.Push(MakeValue(int(vm.stack[vm.sp-1].Kind)))
		},
		OpCLOAD: func() error { return vm.LoadConstant(index) },
		OpLOAD:  func() error { return vm.LoadVariable(index) },

		OpSDUP: func() error {
			if vm.sp < 1 {
				return ErrorUnderflow
			} else if vm.sp == stackSize {
				return ErrorOverflow
			}

			vm.stack[vm.sp] = vm.stack[vm.sp-1]
			vm.sp++
			return nil
		},
		OpSDROP: func() error {
			if vm.sp < 1 {
				return ErrorUnderflow
			}

			vm.sp--
			vm.stack[vm.sp] = ValueNil
			return nil
		},
		OpSSWAP: func() error {
			if vm.sp < 2 {
				return ErrorUnderflow
			}

			vm.stack[vm.sp-1], vm.stack[vm.sp-2] = vm.stack[vm.sp-2], vm.stack[vm.sp-1]
			return nil
		},
		OpSOVER: func() error {
			if vm.sp < 2 {
				return ErrorUnderflow
			} else if vm.sp == stackSize {
				return ErrorOverflow
			}

			vm.stack[vm.sp] = vm.stack[vm.sp-2]
			vm.sp++
			return nil
		},
		OpSROT: func() error {
			if vm.sp < 3 {
				return ErrorUnderflow
			}

			vm.stack[vm.sp-1], vm.stack[vm.sp-2], vm.stack[vm.sp-3] =
				vm.stack[vm.sp-2], vm.stack[vm.sp-3], vm.stack[vm.sp-1]
			return nil
		},

		OpCMP: func() error {
			if vm.sp < 2 {
				return ErrorUnderflow
			}

			vLhs := vm.stack[vm.sp-2]
			vRhs := vm.stack[vm.sp-1]
			vm.sp -= 1
			vm.stack[vm.sp] = ValueNil

			if vLhs.Kind != vRhs.Kind {
				return ErrorType
			}

			vm.stack[vm.sp-1] = MakeValue(util.Spaceship(vLhs.Data, vRhs.Data))
			return nil
		},
		OpLT: func() error {
			if vm.sp < 1 {
				return ErrorUnderflow
			}

			sRes := float64(0)
			bRes := false

			if sRes, bRes = vm.stack[vm.sp-1].Data.(float64); !bRes {
				return ErrorType
			}

			cType := opcode - OpLT
			bRes = util.MatCond(cType == comparisonTypeLt, sRes < 0) &&
				util.MatCond(cType == comparisonTypeLE, sRes <= 0) &&
				util.MatCond(cType == comparisonTypeEq, sRes == 0) &&
				util.MatCond(cType == comparisonTypeGE, sRes >= 0) &&
				util.MatCond(cType == comparisonTypeGt, sRes > 0) &&
				util.MatCond(cType == comparisonTypeNeq, sRes != 0)

			vm.stack[vm.sp-1] = MakeValue(bRes)
			return nil
		},
		OpLTTBLB: func() error {
			if vm.sp < 2 {
				return ErrorUnderflow
			}

			tTable := uint8(0)
			if opcode == OpLTTBLB {
				tTable = argBytes[0]
			} else {
				switch opcode {
				case OpLAND:
					tTable = uint8(util.RelAnd)
					break
				case OpLOR:
					tTable = uint8(util.RelOr)
					break
				case OpLXOR:
					tTable = uint8(util.RelXor)
					break
				}
			}

			if lhs, rhs, err := vm.Pop2Kind(KindBool); err != nil {
				return err
			} else {
				return vm.Push(MakeValue(util.TTableEval(lhs.Data.(bool), rhs.Data.(bool), util.Relation(tTable))))
			}
		},
		OpLTTBLU: func() error {
			tTable := uint8(0)
			if opcode == OpLTTBLU {
				tTable = argBytes[0]
			} else if opcode == OpLNOT {
				tTable = uint8(util.RelNotP)
			}

			if v, err := vm.PopKind(KindBool); err != nil {
				return err
			} else {
				return vm.Push(MakeValue(util.TTableEval(v.Data.(bool), true, util.Relation(tTable))))
			}
		},

		OpNADD:  func() error { return arithHelper(func(lhs, rhs float64) float64 { return lhs + rhs }) },
		OpNSUB:  func() error { return arithHelper(func(lhs, rhs float64) float64 { return lhs - rhs }) },
		OpNMUL:  func() error { return arithHelper(func(lhs, rhs float64) float64 { return lhs * rhs }) },
		OpNDIV:  func() error { return arithHelper(func(lhs, rhs float64) float64 { return lhs / rhs }) },
		OpNFMOD: func() error { return arithHelper(func(lhs, rhs float64) float64 { return math.Mod(lhs, rhs) }) },
		OpNPOW:  func() error { return arithHelper(func(lhs, rhs float64) float64 { return math.Pow(lhs, rhs) }) },
		OpNSHL: func() error {
			if lhs, rhs, err := vm.Pop2(); err != nil {
				return err
			} else {
				lhsV, rhsV := lhs.Data.(float64), rhs.Data.(float64)

				shiftBy := int64(rhsV)
				if shiftBy < 0 {
					return ErrorArithmetic
				}

				newVal := int64(0)

				if opcode == OpNSHL {
					newVal = int64(lhsV) << shiftBy
				} else {
					newVal = int64(lhsV) >> shiftBy
				}

				return vm.Push(MakeValue(newVal))
			}
		},
		OpNSQRT:  func() error { return unaryArithHelper(func(v float64) float64 { return math.Sqrt(v) }) },
		OpNTRUNC: func() error { return unaryArithHelper(func(v float64) float64 { return math.Trunc(v) }) },
		OpNFLOOR: func() error { return unaryArithHelper(func(v float64) float64 { return math.Floor(v) }) },
		OpNCEIL:  func() error { return unaryArithHelper(func(v float64) float64 { return math.Ceil(v) }) },

		OpHLT: func() error {
			vm.halt = true
			return nil
		},
		OpJMP: func() error { return vm.JumpGeneric(index, opcode-OpJMP) },
		OpDJMP: func() error {
			if idxVal, err := vm.PopKind(KindNumber); err != nil {
				return err
			} else {
				return vm.JumpGeneric(uint32(idxVal.Data.(float64)), opcode-OpJMP)
			}
		},
		OpRET: func() error { return vm.Return() },
	}

	opcodeHandlerAliases := map[uint8]uint8{
		OpNCONST_1: OpNCONST_0,
		OpNCONST_2: OpNCONST_0,
		OpBCONST_1: OpBCONST_0,
		OpNSHR:     OpNSHL,
		OpJMPT:     OpJMP,
		OpJMPF:     OpJMP,
		OpCALL:     OpJMP,
		OpDJMPT:    OpDJMP,
		OpDJMPF:    OpDJMP,
		OpDCALL:    OpDJMP,
		OpLE:       OpLT,
		OpEQ:       OpLT,
		OpGE:       OpLT,
		OpGT:       OpLT,
		OpNE:       OpLT,
		OpLAND:     OpLTTBLB,
		OpLOR:      OpLTTBLB,
		OpLXOR:     OpLTTBLB,
		OpLNOT:     OpLTTBLU,
	}

	queryOpcode := opcode
	if alias, aliasOk := opcodeHandlerAliases[opcode]; aliasOk {
		queryOpcode = alias
	}
	if handler, handlerOk := opcodeHandlers[queryOpcode]; handlerOk {
		return handler()
	}
	return ErrorBadOpcode
}

func (vm *VM) DumpNow() {
	fmt.Println("---------------")
	fmt.Printf("Halted, PC, SP: %t, %d, %d\n", vm.halt, vm.pc, vm.sp)
	fmt.Print("Stack         : [")
	for i := 0; i < stackSize && vm.stack[i] != ValueNil; i++ {
		fmt.Print(vm.stack[i], " ")
	}
	fmt.Print("]\n")

	if vm.pc == len(vm.program.Program) {
		fmt.Println("vm.pc == len(vm.program)")
		return
	}

	var err error

	var opcode byte
	if opcode, err = vm.readByte(); err != nil {
		fmt.Println("Error while reading opcode:", err)
		return
	}

	fmt.Printf("Opcode        : %d (%s)\n", opcode, GetOpcodeProperties(opcode).Name)

	argSize := GetOpcodeProperties(opcode).ArgSize
	argBytes := make([]byte, argSize)

	vm.pc++
	if err = vm.readBytes(argBytes, argSize); err != nil {
		fmt.Println("Error while reading argument bytes:", err)
		vm.pc--
		return
	}
	vm.pc--

	if argSize != 0 {
		fmt.Println("Argument bytes:", argBytes)
	}
}
