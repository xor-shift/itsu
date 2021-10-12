package vm

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type ProgramBuilder struct {
	constantPool []interface{}

	reservedConstantIndices map[string]uint32
	reservedConstantKinds   map[string]int

	buffer bytes.Buffer
}

func NewProgramBuilder() *ProgramBuilder {
	b := &ProgramBuilder{
		constantPool:            make([]interface{}, 0),
		reservedConstantIndices: make(map[string]uint32),
		reservedConstantKinds:   make(map[string]int),
		buffer:                  bytes.Buffer{},
	}

	return b
}

func (b *ProgramBuilder) addConstantGeneric(v interface{}) uint32 {
	b.constantPool = append(b.constantPool, v)
	return uint32(len(b.constantPool) - 1)
}

func (b *ProgramBuilder) AddConstantNumber(v float64) uint32 { return b.addConstantGeneric(v) }
func (b *ProgramBuilder) AddConstantString(v string) uint32  { return b.addConstantGeneric(v) }
func (b *ProgramBuilder) AddConstantBool(v bool) uint32      { return b.addConstantGeneric(v) }

func (b *ProgramBuilder) reserveConstantGeneric(name string, kind int) uint32 {
	var idx uint32

	switch kind {
	case KindNumber:
		idx = b.AddConstantNumber(0)
		break
	case KindBool:
		idx = b.AddConstantBool(false)
		break
	case KindString:
		idx = b.AddConstantString("")
		break
	default:
		panic("bad kind")
	}

	b.reservedConstantKinds[name] = kind
	b.reservedConstantIndices[name] = idx

	return idx
}

func (b *ProgramBuilder) ReserveConstantNumber(name string) uint32 {
	return b.reserveConstantGeneric(name, KindNumber)
}
func (b *ProgramBuilder) ReserveConstantBool(name string) uint32 {
	return b.reserveConstantGeneric(name, KindBool)
}
func (b *ProgramBuilder) ReserveConstantString(name string) uint32 {
	return b.reserveConstantGeneric(name, KindString)
}

func (b *ProgramBuilder) emitGeneric(v interface{}) { binary.Write(&b.buffer, binary.LittleEndian, v) }
func (b *ProgramBuilder) emitBytes(v []byte)        { b.buffer.Write(v) }
func (b *ProgramBuilder) EmitByte(v byte)           { b.buffer.WriteByte(v) }

func (b *ProgramBuilder) EmitNCONST(v float64) {
	switch v {
	case 0.:
		b.EmitByte(OpNCONST_0)
		break
	case 1.:
		b.EmitByte(OpNCONST_1)
		break
	case 2.:
		b.EmitByte(OpNCONST_2)
		break
	default:
		b.EmitByte(OpNCONST)
		b.emitGeneric(v)
		break
	}
}

func (b *ProgramBuilder) EmitBCONST(v bool) {
	if v {
		b.EmitByte(OpBCONST_1)
	} else {
		b.EmitByte(OpBCONST_0)
	}
}

func (b *ProgramBuilder) emitLoadGeneric(index uint32, opcode uint8) {
	b.EmitByte(opcode)
	b.emitGeneric(index)
}

func (b *ProgramBuilder) EmitNLOAD(index uint32)   { b.emitLoadGeneric(index, OpNLOAD) }
func (b *ProgramBuilder) EmitBLOAD(index uint32)   { b.emitLoadGeneric(index, OpBLOAD) }
func (b *ProgramBuilder) EmitSTRLOAD(index uint32) { b.emitLoadGeneric(index, OpSTRLOAD) }

func (b *ProgramBuilder) MapConstants(m map[string]interface{}) (err error) {
	for key, kind := range b.reservedConstantKinds {
		var v interface{}
		var ok, tok bool

		v, ok = m[key]

		switch kind {
		case KindNumber:
			_, tok = v.(float64)
			break
		case KindBool:
			_, tok = v.(bool)
			break
		case KindString:
			_, tok = v.(string)
			break
		}

		if !ok {
			err = errors.New("missing constant in constant mapping")
		} else if !tok {
			err = errors.New("bad constant type in constant mapping")
		}

		b.constantPool[b.reservedConstantIndices[key]] = v
	}

	return
}
