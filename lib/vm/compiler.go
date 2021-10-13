package vm

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type ProgramBuilder struct {
	constantPool []Value

	reservedConstantIndices map[string]uint32

	buffer bytes.Buffer
}

func NewProgramBuilder() *ProgramBuilder {
	b := &ProgramBuilder{
		constantPool:            make([]Value, 0),
		reservedConstantIndices: make(map[string]uint32),
		buffer:                  bytes.Buffer{},
	}

	return b
}

func (b *ProgramBuilder) AddConstant(v Value) uint32 {
	b.constantPool = append(b.constantPool, v)
	return uint32(len(b.constantPool) - 1)
}

func (b *ProgramBuilder) ReserveConstant(name string) uint32 {
	if existingIdx, ok := b.reservedConstantIndices[name]; ok {
		return existingIdx
	} else {
		idx := b.AddConstant(MakeValue(nil))

		b.reservedConstantIndices[name] = idx

		return idx
	}
}

func (b *ProgramBuilder) emitGeneric(v interface{}) { binary.Write(&b.buffer, binary.LittleEndian, v) }
func (b *ProgramBuilder) emitBytes(v []byte)        { b.buffer.Write(v) }
func (b *ProgramBuilder) EmitByte(v byte)           { b.buffer.WriteByte(v) }

func (b *ProgramBuilder) EmitCLoad(index uint32) {
	b.EmitByte(OpCLOAD)
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
		b.EmitCLoad(b.AddConstant(v))
		break
	}
}

type BuiltProgram struct {
	program []byte

	constantPool []Value

	reservedConstantIndices map[string]uint32
}

func (b BuiltProgram) GobEncode() ([]byte, error) {
	return b.Serialize()
}

func (b *BuiltProgram) GobDecode(data []byte) error {
	if b2, err := DeserializeBuiltProgram(bufio.NewReader(bytes.NewReader(data))); err != nil {
		return err
	} else {
		b.program = b2.program
		b.constantPool = b2.constantPool
		b.reservedConstantIndices = b2.reservedConstantIndices
		return nil
	}
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
	}

	buf = buffer.Bytes()

	return
}

func DeserializeBuiltProgram(reader *bufio.Reader) (b BuiltProgram, err error) {
	b = BuiltProgram{
		program:                 nil,
		constantPool:            nil,
		reservedConstantIndices: make(map[string]uint32),
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

		b.reservedConstantIndices[string(keyBuffer)] = index
	}

	return
}

func (b *ProgramBuilder) Build() BuiltProgram {
	b2 := BuiltProgram{
		program:                 b.buffer.Bytes(),
		constantPool:            b.constantPool,
		reservedConstantIndices: b.reservedConstantIndices,
	}

	b.buffer = bytes.Buffer{}
	b.constantPool = make([]Value, 0)
	b.reservedConstantIndices = make(map[string]uint32)

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

	for key, index := range b.reservedConstantIndices {
		if v, ok := m[key]; !ok {
			if err = errors.New(fmt.Sprint("missing constant in constant mapping with key: ", key)); err != nil {
				return
			}
		} else {
			b.constantPool[index] = MakeValue(v)
		}
	}

	return
}
