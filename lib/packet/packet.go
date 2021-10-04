package packet

import (
	"bufio"
	"crypto/ed25519"
	itsu_crpyto "example.com/itsuMain/lib/crpyto"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"io"
)

type Packet struct {
	SignatureType itsu_crpyto.SigType
	Signature     []byte
	Data          []byte
}

func NewPacket(data []byte) Packet {
	m := Packet{
		SignatureType: itsu_crpyto.SigTypeNone,
		Signature:     make([]byte, 0),
		Data:          data,
	}

	return m
}

//SerializeTo will try to serialize a Packet. Some data might have been written if an error is returned. The data will be attempted to be compressed.
func (packet Packet) SerializeTo(writer io.Writer) (n int, err error) {
	tempN := 0

	header := Header{
		SignatureType: packet.SignatureType,
		UCSize:        0,
		PayloadSize:   uint64(len(packet.Data)),
		Signature:     packet.Signature,
	}

	var payload []byte
	if payload, err = util.TryCompress(packet.Data); err == nil {
		header.UCSize = header.PayloadSize
		header.PayloadSize = uint64(len(payload))
	}

	if tempN, err = header.SerializeTo(writer); err != nil {
		return
	}
	n += tempN

	if tempN, err = writer.Write(payload); err != nil {
		return
	}
	n += tempN

	return
}

func (packet *Packet) DeserializeFrom(reader *bufio.Reader) (err error) {
	header := Header{}
	if err = header.DeserializeFrom(reader); err != nil {
		return
	}

	payload := make([]byte, header.PayloadSize)
	if _, err = io.ReadFull(reader, payload); err != nil {
		return
	}

	packet.SignatureType = header.SignatureType
	packet.Signature = header.Signature

	if header.IsCompressed() {
		var uncompressed []byte
		if uncompressed, err = util.TryDecompress(payload, int64(header.UCSize)); err != nil {
			return
		}

		packet.Data = uncompressed
	} else {
		packet.Data = payload
	}

	return
}

func (packet *Packet) PreSign(signatureType itsu_crpyto.SigType, signature []byte) error {
	if len(signature) != itsu_crpyto.SignatureSize(signatureType) {
		return ErrorHeaderBadSignatureSize
	}

	packet.SignatureType = signatureType
	packet.Signature = signature

	return nil
}

func (packet *Packet) SignED25519(key ed25519.PrivateKey) (err error) {
	return packet.PreSign(itsu_crpyto.SigTypeED25519, ed25519.Sign(key, packet.Data))
}

func (packet *Packet) VerifySignature() error {
	return itsu_crpyto.VerifyClientSignature(packet.Data, packet.Signature, packet.SignatureType)
}

func (packet *Packet) VerifySignatureFull(m message.Msg, expectedToken uint64) (err error) {
	if sm, ok := m.(message.SignedMessage); ok {
		if sm.GetSignatureToken() != expectedToken {
			return itsu_crpyto.ErrorClientSigBadToken
		}
		return packet.VerifySignature()
	} else {
		return itsu_crpyto.ErrorClientSigInternal
	}
}
