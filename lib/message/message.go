package message

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"example.com/itsuMain/lib/util"
	"io"
)

type Msg interface {
	GetID() MessageID
}

type SignedMessage interface {
	Msg
	GetSignatureToken() uint64
}

type SignableMessage interface {
	SignedMessage
	SetSignatureToken(uint64)
}

func SerializeMessageTo(writer io.Writer, m Msg) (err error) {
	encoder := gob.NewEncoder(writer)

	if _, err = util.WriteUvarint(writer, uint64(m.GetID())); err != nil {
		return
	}

	err = encoder.Encode(&m)

	return
}

func SerializeMessage(m Msg) (data []byte) {
	var buf bytes.Buffer

	_ = SerializeMessageTo(&buf, m)
	data = buf.Bytes()

	return
}

func DeserializeMessageFrom(reader *bufio.Reader) (m Msg, err error) {
	var messageID uint64
	if messageID, err = binary.ReadUvarint(reader); err != nil {
		return
	}
	messageID &= messageID //uhh

	decoder := gob.NewDecoder(reader)
	err = decoder.Decode(&m)
	return
}

func DeserializeMessage(data []byte) (m Msg, err error) {
	buf := bytes.NewBuffer(data)
	return DeserializeMessageFrom(bufio.NewReader(buf))
}
