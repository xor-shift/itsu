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

type VM struct {
	program Program
	pc      int

	stack [stackSize]Value
	sp    int

	callStack [callStackSize]uint32
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

func (vm *VM) readByte() (byte, error) {
	if vm.pc == len(vm.program.Program) {
		return 0, ErrorBadEOF
	}

	b := vm.program.Program[vm.pc]
	return b, nil
}

func (vm *VM) readBytes(buf []byte, n int) error {
	if vm.pc+n > len(vm.program.Program) {
		return ErrorBadEOF
	}

	for i := 0; i < n; i++ {
		buf[i] = vm.program.Program[vm.pc+i]
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
	slv.Index(*ptr).Set(reflect.ValueOf(nil))

	return val, nil
}

func (vm *VM) Push(a Value) error { return genericPush(&vm.stack, &vm.sp, a) }

func (vm *VM) Top() (Value, error) {
	if v, err := genericTop(vm.stack, vm.sp); err != nil {
		return ValueNil, ErrorUnderflow
	} else {
		return v.(Value), nil
	}
}

func (vm *VM) Pop() (Value, error) {
	if v, err := genericPop(&vm.stack, &vm.sp); err != nil {
		return ValueNil, ErrorUnderflow
	} else {
		return v.(Value), nil
	}
}

func (vm *VM) PopKind(kind Kind) (Value, error) {
	if v, err := vm.Pop(); err != nil {
		return ValueNil, err
	} else {
		if v.Kind != kind {
			return ValueNil, ErrorType
		}
		return v, nil
	}
}

func (vm *VM) CPush(a uint32) error { return genericPush(&vm.callStack, &vm.csp, a) }

func (vm *VM) CTop() (uint32, error) {
	if v, err := genericTop(vm.callStack, vm.csp); err != nil {
		return 0, ErrorUnderflow
	} else {
		return v.(uint32), nil
	}
}

func (vm *VM) CPop() (uint32, error) {
	if v, err := genericPop(&vm.callStack, &vm.csp); err != nil {
		return 0, ErrorUnderflow
	} else {
		return v.(uint32), nil
	}
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

	binaryNOp := func(fn func(lhs, rhs float64) (interface{}, error)) error {
		if vm.sp < 2 {
			return ErrorUnderflow
		}

		var lhs, rhs float64
		var ok bool

		if lhs, ok = vm.stack[vm.sp-2].Data.(float64); !ok {
			return ErrorType
		}

		if rhs, ok = vm.stack[vm.sp-1].Data.(float64); !ok {
			return ErrorType
		}

		vm.sp--
		vm.stack[vm.sp] = ValueNil

		var tData interface{}
		tData, err = fn(lhs, rhs)
		vm.stack[vm.sp-1] = MakeValue(tData)

		return err
	}

	unaryNOp := func(fn func(v float64) (interface{}, error)) error {
		if vm.sp < 1 {
			return ErrorUnderflow
		}

		var v float64
		var ok bool
		if v, ok = vm.stack[vm.sp-1].Data.(float64); !ok {
			return ErrorType
		}

		var tData interface{}
		tData, err = fn(v)
		vm.stack[vm.sp-1] = MakeValue(tData)
		return err
	}

	//types are: jmp, jmpt, jmpf, call
	doJMPS := func(to uint32, typ uint8) (err error) {
		doJmp := false

		if typ == 0 || typ == 3 {
			doJmp = true
		} else {
			if vm.sp < 1 {
				return ErrorUnderflow
			}

			val := vm.stack[vm.sp-1]
			if val.Kind != KindBool {
				return ErrorType
			}

			doJmp = (typ == 1) == (val.Data.(bool) == true)
		}

		if !doJmp {
			return nil
		}

		if typ == 3 {
			if err = vm.CPush(uint32(vm.pc)); err != nil {
				return
			}
		}

		vm.sp = int(index)

		return
	}

	switch opcode {
	case OpNCONST, OpNCONST_0, OpNCONST_1, OpNCONST_2:
		if opcode == OpNCONST {
			err = vm.Push(MakeValue(number))
		} else {
			err = vm.Push(MakeValue(opcode - OpNCONST_0))
		}

		if err != nil {
			return
		}

		break

	case OpBCONST_0, OpBCONST_1:
		if err = vm.Push(MakeValue(opcode == OpBCONST_1)); err != nil {
			return
		}
		break

	case OpNLOAD, OpBLOAD, OpSTRLOAD:
		loaded := vm.program.Constants[index]

		bad := false
		switch loaded.Data.(type) {
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

		if err = vm.Push(loaded); err != nil {
			return
		}

		break

	case OpNILCONST:
		if err = vm.Push(ValueNil); err != nil {
			return
		}
		break

	case OpISNIL:
		if vm.sp == 0 {
			return ErrorUnderflow
		}

		if err = vm.Push(MakeValue(vm.stack[vm.sp-1].Kind == KindNil)); err != nil {
			return
		}

		break

	case OpKIND:
		if vm.sp == 0 {
			return ErrorUnderflow
		}

		if err = vm.Push(MakeValue(int(vm.stack[vm.sp-1].Kind))); err != nil {
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

		vm.sp--
		vm.stack[vm.sp] = ValueNil
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
		vm.stack[vm.sp] = ValueNil

		if reflect.TypeOf(vLhs).Kind() != reflect.TypeOf(vRhs).Kind() {
			return ErrorType
		}

		vm.stack[vm.sp-1] = MakeValue(util.Spaceship(vLhs, vRhs))
		break

	case OpLT, OpLE, OpEQ, OpGE, OpGT, OpNE:
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

		if lhs, ok = vm.stack[vm.sp-2].Data.(bool); !ok {
			return ErrorType
		}
		if rhs, ok = vm.stack[vm.sp-1].Data.(bool); !ok {
			return ErrorType
		}

		vm.sp--
		vm.stack[vm.sp] = ValueNil
		vm.stack[vm.sp-1] = MakeValue(util.TTableEval(lhs, rhs, util.Relation(tTable)))
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
		if p, ok = vm.stack[vm.sp-1].Data.(bool); !ok {
			return ErrorType
		}

		vm.stack[vm.sp-1] = MakeValue(util.TTableEval(p, true, util.Relation(tTable)))

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

	case OpHLT:
		vm.halt = true
		break

	case OpJMP, OpJMPT, OpJMPF, OpCALL:
		if err = doJMPS(index, opcode-OpJMP); err != nil {
			return
		}
		break

	case OpDJMP, OpDJMPT, OpDJMPF, OpDCALL:
		var idxVal Value
		if idxVal, err = vm.PopKind(KindNumber); err != nil {
			return
		}

		if err = doJMPS(uint32(idxVal.Data.(float64)), opcode-OpJMP); err != nil {
			return
		}
		break

	case OpRET:
		var retPoint uint32
		if retPoint, err = vm.CPop(); err != nil {
			return
		}

		vm.pc = int(retPoint)

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
