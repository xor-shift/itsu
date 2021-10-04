package util

import (
	"encoding/binary"
	"io"
)

type VarWriter struct {
	buf []byte
}

func NewVarWriter() VarWriter {
	return VarWriter{
		buf: make([]byte, binary.MaxVarintLen64),
	}
}

func (w VarWriter) WriteVarint(writer io.Writer, v int64) (n int, err error) {
	return writer.Write(w.buf[:binary.PutVarint(w.buf, v)])
}

func (w VarWriter) WriteUvarint(writer io.Writer, v uint64) (n int, err error) {
	return writer.Write(w.buf[:binary.PutUvarint(w.buf, v)])
}

func WriteUvarint(writer io.Writer, v uint64) (n int, err error) {
	wr := NewVarWriter()
	return wr.WriteUvarint(writer, v)
}

func WriteVarint(writer io.Writer, v int64) (n int, err error) {
	wr := NewVarWriter()
	return wr.WriteVarint(writer, v)
}
