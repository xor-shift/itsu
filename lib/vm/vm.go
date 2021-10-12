package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"example.com/itsuMain/lib/util"
	"fmt"
	"math"
	"reflect"
)

const (
	stackSize = 64
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

type VM struct {
	constants []interface{}

	program []byte
	pc      int

	stack [stackSize]interface{}
	sp    int

	halt bool
}

func NewVM() *VM {
	v := &VM{
		constants: make([]interface{}, 0),
		program:   nil,
		pc:        0,
		stack:     [stackSize]interface{}{},
		sp:        0,
	}

	return v
}

func (vm *VM) LoadFromBuilder(b *ProgramBuilder) {
	vm.program = b.buffer.Bytes()
	vm.constants = b.constantPool

	b.constantPool = make([]interface{}, 0)
	b.buffer = bytes.Buffer{}
}

func (vm *VM) readByte() (byte, error) {
	if vm.pc == len(vm.program) {
		return 0, ErrorBadEOF
	}

	b := vm.program[vm.pc]
	return b, nil
}

func (vm *VM) readBytes(buf []byte, n int) error {
	if vm.pc+n > len(vm.program) {
		return ErrorBadEOF
	}

	for i := 0; i < n; i++ {
		buf[i] = vm.program[vm.pc+i]
	}

	return nil
}

func (vm *VM) readInstruction() (opcode byte, argBytes []byte, err error) {
	if opcode, err = vm.readByte(); err != nil {
		return
	}

	props := GetOpcodeProperties(opcode)
	if props.Bad() {
		err = ErrorBadOpcode
		return
	}
	vm.pc += 1

	argBytes = make([]byte, props.ArgSize)
	if err = vm.readBytes(argBytes, props.ArgSize); err != nil {
		return
	}
	vm.pc += props.ArgSize

	return
}

func (vm *VM) push(a interface{}) error {
	if vm.sp == stackSize {
		return ErrorOverflow
	}

	vm.stack[vm.sp] = a
	vm.sp++

	return nil
}

func (vm *VM) top() (interface{}, error) {
	if vm.sp == 0 {
		return nil, ErrorUnderflow
	}

	return vm.stack[vm.sp-1], nil
}

func (vm *VM) pop() (interface{}, error) {
	if v, err := vm.top(); err != nil {
		return nil, err
	} else {
		vm.sp--
		vm.stack[vm.sp] = nil
		return v, nil
	}
}

func (vm *VM) SingleStep() (err error) {
	if vm.pc == len(vm.program) {
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

	binaryNOp := func(fn func(lhs, rhs float64) (interface{}, error)) error {
		if vm.sp < 2 {
			return ErrorUnderflow
		}

		var lhs, rhs float64
		var ok bool

		if lhs, ok = vm.stack[vm.sp-2].(float64); !ok {
			return ErrorType
		}

		if rhs, ok = vm.stack[vm.sp-1].(float64); !ok {
			return ErrorType
		}

		vm.sp--
		vm.stack[vm.sp] = nil
		vm.stack[vm.sp-1], err = fn(lhs, rhs)

		return err
	}

	unaryNOp := func(fn func(v float64) (interface{}, error)) error {
		if vm.sp < 1 {
			return ErrorUnderflow
		}

		var v float64
		var ok bool
		if v, ok = vm.stack[vm.sp-1].(float64); !ok {
			return ErrorType
		}

		vm.stack[vm.sp-1], err = fn(v)
		return err
	}

	switch opcode {
	case OpNCONST, OpNCONST_0, OpNCONST_1, OpNCONST_2:
		var narg float64

		if opcode == OpNCONST {
			narg = number
		} else {
			narg = float64(opcode - OpNCONST_0)
		}

		if err = vm.push(narg); err != nil {
			return
		}

		break

	case OpBCONST_0, OpBCONST_1:
		if err = vm.push(opcode == OpBCONST_1); err != nil {
			return
		}
		break

	case OpNLOAD, OpBLOAD, OpSTRLOAD:
		loaded := vm.constants[index]

		bad := false
		switch loaded.(type) {
		case float64:
			bad = opcode != OpNLOAD
			break
		case bool:
			bad = opcode != OpBLOAD
			break
		case string:
			bad = opcode != OpSTRLOAD
			break
		default:
			return ErrorUnsupportedConstant
		}

		if bad {
			return ErrorType
		}

		if err = vm.push(loaded); err != nil {
			return
		}

		break

	case OpSDUP:
		if vm.sp < 1 {
			return ErrorUnderflow
		} else if vm.sp == stackSize {
			return ErrorOverflow
		}

		vm.stack[vm.sp] = vm.stack[vm.sp-1]
		vm.sp++
		break

	case OpSDROP:
		if vm.sp < 1 {
			return ErrorUnderflow
		}

		vm.stack[vm.sp-1] = nil
		vm.sp--
		break

	case OpSSWAP:
		if vm.sp < 2 {
			return ErrorUnderflow
		}

		vm.stack[vm.sp-1], vm.stack[vm.sp-2] = vm.stack[vm.sp-2], vm.stack[vm.sp-1]
		break

	case OpSOVER:
		if vm.sp < 2 {
			return ErrorUnderflow
		} else if vm.sp == stackSize {
			return ErrorOverflow
		}

		vm.stack[vm.sp] = vm.stack[vm.sp-2]
		vm.sp++
		break

	case OpSROT:
		if vm.sp < 3 {
			return ErrorUnderflow
		}

		vm.stack[vm.sp-1], vm.stack[vm.sp-2], vm.stack[vm.sp-3] =
			vm.stack[vm.sp-2], vm.stack[vm.sp-3], vm.stack[vm.sp-1]
		break

	case OpNCMP, OpSCMP:
		if vm.sp < 2 {
			return ErrorUnderflow
		}

		vLhs := vm.stack[vm.sp-2]
		vRhs := vm.stack[vm.sp-1]
		vm.sp -= 1
		vm.stack[vm.sp] = nil

		if reflect.TypeOf(vLhs).Kind() != reflect.TypeOf(vRhs).Kind() {
			return ErrorType
		}

		res := util.Spaceship(vLhs, vRhs)
		vm.stack[vm.sp-1] = float64(res)
		break

	case OpLT, OpLE, OpEQ, OpGE, OpGT, OpNE:
		if vm.sp < 1 {
			return ErrorUnderflow
		}

		sRes := float64(0)
		bRes := false

		if sRes, bRes = vm.stack[vm.sp-1].(float64); !bRes {
			return ErrorType
		}

		cType := opcode - OpLT
		bRes = util.MatCond(cType == comparisonTypeLt, sRes < 0) &&
			util.MatCond(cType == comparisonTypeLE, sRes <= 0) &&
			util.MatCond(cType == comparisonTypeEq, sRes == 0) &&
			util.MatCond(cType == comparisonTypeGE, sRes >= 0) &&
			util.MatCond(cType == comparisonTypeGt, sRes > 0) &&
			util.MatCond(cType == comparisonTypeNeq, sRes != 0)

		vm.stack[vm.sp-1] = bRes
		break

	case OpLAND, OpLOR, OpLXOR, OpLTTBLB:
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

		var lhs, rhs bool
		var ok bool

		if lhs, ok = vm.stack[vm.sp-2].(bool); !ok {
			return ErrorType
		}
		if rhs, ok = vm.stack[vm.sp-1].(bool); !ok {
			return ErrorType
		}

		vm.sp--
		vm.stack[vm.sp] = nil
		vm.stack[vm.sp-1] = util.TTableEval(lhs, rhs, util.Relation(tTable))
		break

	case OpLNOT, OpLTTBLU:
		if vm.sp < 1 {
			return ErrorUnderflow
		}

		tTable := uint8(0)
		if opcode == OpLTTBLU {
			tTable = argBytes[0]
		} else if opcode == OpLNOT {
			tTable = uint8(util.RelNotP)
		}

		var p, ok bool
		if p, ok = vm.stack[vm.sp-1].(bool); !ok {
			return ErrorType
		}

		vm.stack[vm.sp-1] = util.TTableEval(p, true, util.Relation(tTable))

		break

	case OpHLT:
		vm.halt = true
		break

	case OpNADD, OpNSUB, OpNMUL, OpNDIV, OpNFMOD, OpNPOW, OpNSHL, OpNSHR:
		fns := map[uint8]func(lhs, rhs float64) (interface{}, error){
			OpNADD:  func(lhs, rhs float64) (interface{}, error) { return lhs + rhs, nil },
			OpNSUB:  func(lhs, rhs float64) (interface{}, error) { return lhs - rhs, nil },
			OpNMUL:  func(lhs, rhs float64) (interface{}, error) { return lhs * rhs, nil },
			OpNDIV:  func(lhs, rhs float64) (interface{}, error) { return lhs / rhs, nil },
			OpNFMOD: func(lhs, rhs float64) (interface{}, error) { return math.Mod(lhs, rhs), nil },
			OpNPOW:  func(lhs, rhs float64) (interface{}, error) { return math.Pow(lhs, rhs), nil },
			OpNSHL: func(lhs, rhs float64) (interface{}, error) {
				shiftBy := int64(rhs)
				if shiftBy < 0 {
					return nil, ErrorArithmetic
				}

				return int64(lhs) << shiftBy, nil
			},
			OpNSHR: func(lhs, rhs float64) (interface{}, error) {
				shiftBy := int64(rhs)
				if shiftBy < 0 {
					return nil, ErrorArithmetic
				}

				return int64(lhs) >> shiftBy, nil
			},
		}

		err = binaryNOp(fns[opcode])
		break

	case OpNSQRT, OpNTRUNC, OpNFLOOR, OpNCEIL:
		fns := map[uint8]func(v float64) (interface{}, error){
			OpNSQRT:  func(v float64) (interface{}, error) { return math.Sqrt(v), nil },
			OpNTRUNC: func(v float64) (interface{}, error) { return math.Trunc(v), nil },
			OpNFLOOR: func(v float64) (interface{}, error) { return math.Floor(v), nil },
			OpNCEIL:  func(v float64) (interface{}, error) { return math.Ceil(v), nil },
		}

		err = unaryNOp(fns[opcode])
		break

	default:
		err = ErrorBadOpcode
		break
	}

	return
}

func (vm *VM) DumpNow() {
	fmt.Println("---------------")
	fmt.Printf("Halted, PC, SP: %t, %d, %d\n", vm.halt, vm.pc, vm.sp)
	fmt.Print("Stack         : [")
	for i := 0; i < stackSize && vm.stack[i] != nil; i++ {
		fmt.Print(vm.stack[i], " ")
	}
	fmt.Print("]\n")

	if vm.pc == len(vm.program) {
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
	if err = vm.readBytes(argBytes, argSize); err != nil {
		fmt.Println("Error while reading argument bytes:", err)
		return
	}

	if argSize != 0 {
		fmt.Println("Argument bytes:", argBytes)
	}
}
