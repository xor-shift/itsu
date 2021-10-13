package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type ProgramBuilder struct {
	constantPool []Value

	reservedConstantIndices map[string]uint32
	reservedConstantKinds   map[string]Kind

	buffer bytes.Buffer
}

func NewProgramBuilder() *ProgramBuilder {
	b := &ProgramBuilder{
		constantPool:            make([]Value, 0),
		reservedConstantIndices: make(map[string]uint32),
		reservedConstantKinds:   make(map[string]Kind),
		buffer:                  bytes.Buffer{},
	}

	return b
}

func (b *ProgramBuilder) AddConstant(v Value) uint32 {
	for k, v2 := range b.constantPool {
		if v == v2 {
			return uint32(k)
		}
	}

	b.constantPool = append(b.constantPool, v)
	return uint32(len(b.constantPool) - 1)
}

func (b *ProgramBuilder) ReserveConstant(name string, kind Kind) uint32 {
	if existingIdx, ok := b.reservedConstantIndices[name]; ok {
		return existingIdx
	} else {
		idx := b.AddConstant(ZeroValue(kind))

		b.reservedConstantIndices[name] = idx
		b.reservedConstantKinds[name] = kind

		return idx
	}
}

func (b *ProgramBuilder) emitGeneric(v interface{}) { binary.Write(&b.buffer, binary.LittleEndian, v) }
func (b *ProgramBuilder) emitBytes(v []byte)        { b.buffer.Write(v) }
func (b *ProgramBuilder) EmitByte(v byte)           { b.buffer.WriteByte(v) }

func (b *ProgramBuilder) EmitLoad(index uint32, kind Kind) {
	switch kind {
	case KindNumber:
		b.EmitByte(OpNLOAD)
		break
	case KindBool:
		b.EmitByte(OpBLOAD)
		break
	case KindString:
		b.EmitByte(OpSTRLOAD)
		break
	default:
		return
	}
	b.emitGeneric(index)
}

func (b *ProgramBuilder) EmitConst(v Value) {
	switch v.Kind {
	case KindNumber:
		if n := v.Data.(float64); n == 0. {
			b.EmitByte(OpNCONST_0)
		} else if n == 1. {
			b.EmitByte(OpNCONST_1)
		} else if n == 2. {
			b.EmitByte(OpNCONST_2)
		} else {
			b.EmitByte(OpNCONST)
			b.emitGeneric(n)
		}
		break
	case KindBool:
		if v.Data.(bool) {
			b.EmitByte(OpBCONST_1)
		} else {
			b.EmitByte(OpBCONST_0)
		}
		break
	case KindString:
		b.EmitLoad(b.AddConstant(v), KindString)
		break
	}
}

type BuiltProgram struct {
	program []byte

	constantPool []Value

	reservedConstantIndices map[string]uint32
	reservedConstantKinds   map[string]Kind
}

func (b *ProgramBuilder) Build() BuiltProgram {
	b2 := BuiltProgram{
		program:                 b.buffer.Bytes(),
		constantPool:            b.constantPool,
		reservedConstantIndices: b.reservedConstantIndices,
		reservedConstantKinds:   b.reservedConstantKinds,
	}

	b.buffer = bytes.Buffer{}
	b.constantPool = make([]Value, 0)
	b.reservedConstantIndices = make(map[string]uint32)
	b.reservedConstantKinds = make(map[string]Kind)

	return b2
}

type Program struct {
	Program   []byte
	Constants []Value
}

func (b BuiltProgram) Link(m map[string]interface{}) (p Program, err error) {
	p = Program{
		Program:   b.program,
		Constants: b.constantPool,
	}

	for key, kind := range b.reservedConstantKinds {
		if v, ok := m[key]; !ok {
			if err = errors.New("missing constant in constant mapping"); err != nil {
				return
			}
		} else {
			val := MakeValue(v)
			if val.Kind != kind {
				if err = errors.New("bad given type in constant mapping"); err != nil {
					return
				}
			}
			b.constantPool[b.reservedConstantIndices[key]] = MakeValue(v)
		}
	}

	return
}
