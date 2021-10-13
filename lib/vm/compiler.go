package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
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

func (b BuiltProgram) Serialize() (buf []byte, err error) {
	buffer := bytes.Buffer{}

	if err = binary.Write(&buffer, binary.LittleEndian, uint32(len(b.program))); err != nil {
		return
	}
	buffer.Write(b.program)

	if err = binary.Write(&buffer, binary.LittleEndian, uint32(len(b.constantPool))); err != nil {
		return
	}
	for _, v := range b.constantPool {
		buffer.Write(v.Serialize())
	}

	if err = binary.Write(&buffer, binary.LittleEndian, uint32(len(b.reservedConstantIndices))); err != nil {
		return
	}
	for k, v := range b.reservedConstantIndices {
		if err = binary.Write(&buffer, binary.LittleEndian, uint32(len(k))); err != nil {
			return
		}
		buffer.Write([]byte(k))
		if err = binary.Write(&buffer, binary.LittleEndian, v); err != nil {
			return
		}
		if err = binary.Write(&buffer, binary.LittleEndian, uint16(b.reservedConstantKinds[k])); err != nil {
			return
		}
	}

	buf = buffer.Bytes()

	return
}

func DeserializeBuiltProgram(reader *bufio.Reader) (b BuiltProgram, err error) {
	b = BuiltProgram{
		program:                 nil,
		constantPool:            nil,
		reservedConstantIndices: make(map[string]uint32),
		reservedConstantKinds:   make(map[string]Kind),
	}

	var progLen uint32
	if err = binary.Read(reader, binary.LittleEndian, &progLen); err != nil {
		return
	}
	b.program = make([]byte, progLen)
	if _, err = io.ReadFull(reader, b.program); err != nil {
		return
	}

	var constsLen uint32
	if err = binary.Read(reader, binary.LittleEndian, &constsLen); err != nil {
		return
	}
	b.constantPool = make([]Value, constsLen)
	for i := uint32(0); i < constsLen; i++ {
		var val Value
		if val, err = DeserializeValue(reader); err != nil {
			return
		}

		b.constantPool[i] = val
	}

	var reservedLen uint32
	if err = binary.Read(reader, binary.LittleEndian, &reservedLen); err != nil {
		return
	}
	for i := uint32(0); i < reservedLen; i++ {
		var keyLength uint32
		var keyBuffer []byte
		var index uint32
		var kind uint16

		if err = binary.Read(reader, binary.LittleEndian, &keyLength); err != nil {
			return
		}

		keyBuffer = make([]byte, keyLength)
		if _, err = io.ReadFull(reader, keyBuffer); err != nil {
			return
		}

		if err = binary.Read(reader, binary.LittleEndian, &index); err != nil {
			return
		}

		if err = binary.Read(reader, binary.LittleEndian, &kind); err != nil {
			return
		}

		b.reservedConstantIndices[string(keyBuffer)] = index
		b.reservedConstantKinds[string(keyBuffer)] = Kind(kind)

	}

	return
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
