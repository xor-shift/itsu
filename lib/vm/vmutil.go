package vm

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

func (vm *VM) PopKind(kind Kind) (v Value, err error) {
	if v, err = vm.Pop(); err == nil && v.Kind != kind {
		err = ErrorType
	}

	return
}

//returns the second and the first element on the stack in order
func (vm *VM) Pop2() (lhs, rhs Value, err error) {
	if lhs, err = vm.Pop(); err != nil {
		return
	}

	if rhs, err = vm.Pop(); err != nil {
		return
	}

	return
}

func (vm *VM) Pop2Kinds(kLhs, kRhs Kind) (lhs, rhs Value, err error) {
	if lhs, rhs, err = vm.Pop2(); err != nil {
		return
	}

	if lhs.Kind != kLhs || rhs.Kind != kRhs {
		err = ErrorType
	}

	return
}

func (vm *VM) Pop2Kind(kind Kind) (lhs, rhs Value, err error) { return vm.Pop2Kinds(kind, kind) }

func (vm *VM) Goto(index uint32) {
	vm.pc = int(index)
}

func (vm *VM) Call(pc uint32) error {
	if err := genericPush(&vm.callStack, &vm.csp, callFrame{
		rp:   vm.pc,
		vars: make(map[uint32]Value),
	}); err != nil {
		return err
	}

	vm.pc = int(pc)
	return nil
}

func (vm *VM) Return() error {
	if cfEFace, err := genericPop(&vm.callStack, &vm.csp); err != nil {
		return err
	} else {
		cf := cfEFace.(callFrame)
		vm.sp = cf.rp
		return nil
	}
}

func (vm *VM) StoreVariable(idx uint32, val Value) error {
	if vm.csp < 1 {
		return ErrorUnderflow
	}

	vm.callStack[vm.csp-1].vars[idx] = val
	return nil
}

func (vm *VM) LoadVariable(idx uint32) error {
	if vm.csp >= callStackSize {
		return ErrorUnderflow
	}

	if val, ok := vm.callStack[vm.csp-1].vars[idx]; !ok {
		return vm.Push(ValueNil)
	} else {
		return vm.Push(val)
	}
}

func (vm *VM) LoadConstant(idx uint32) error {
	if vm.csp >= stackSize {
		return ErrorUnderflow
	}

	if idx >= uint32(len(vm.program.Constants)) {
		return vm.Push(ValueNil)
	} else {
		return vm.Push(vm.program.Constants[idx])
	}
}

//JumpGenericfunc types are: jmp, jmpt, jmpf, call
func (vm *VM) JumpGeneric(to uint32, typ uint8) (err error) {
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
		if err = vm.Call(to); err != nil {
			return
		}
	} else {
		vm.Goto(to)
	}

	return
}
