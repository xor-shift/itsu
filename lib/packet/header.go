package packet

import (
	"bufio"
	"encoding/binary"
	"errors"
	itsu_crpyto "example.com/itsuMain/lib/crpyto"
	"example.com/itsuMain/lib/util"
	"io"
)

const (
	MaxDataSize = 1024 * 1024 * 1

	headerFlagsBitCompressed     = 1 << 15
	headerFlagsBitsSignatureType = 0b111 << 12
)

var (
	ErrorHeaderBadSignatureSize   = errors.New("bad signature size")
	ErrorHeaderBadSignatureType   = errors.New("bad signature type")
	ErrorHeaderBadCompressionInfo = errors.New("bad compression information")
	ErrorHeaderLargePayload       = errors.New("large payload")
)

/*
Header format:
[flags (2 bytes)]
[uvarint uncompressed size (1..10 bytes, ignored if compressed == 0)]
[uvarint payload size (1..10 bytes)]
<signature (? bytes)>

flags format:
csss 0000
0000 0000
c -> compression bit // deprecated, nonzero ucsize means compression
sss -> signature type (0 means unsigned message, skip reading the signature part of the header)
	-> 001: ed25519 signature
	-> rest is reserved
*/
type Header struct {
	SignatureType itsu_crpyto.SigType

	UCSize      uint64
	PayloadSize uint64

	Signature []byte
}

func (header Header) IsCompressed() bool {
	return header.UCSize != 0
}

func (header Header) GetValidator() util.Validator {
	return util.Validator{
		{func() bool {
			return itsu_crpyto.SignatureSize(header.SignatureType) == len(header.Signature)
		}, ErrorHeaderBadSignatureSize},
		{func() bool {
			return header.SignatureType <= itsu_crpyto.SigTypeMax
		}, ErrorHeaderBadSignatureType},
		{func() bool {
			return !(header.IsCompressed() && header.PayloadSize >= header.UCSize)
		}, ErrorHeaderBadCompressionInfo},
		{func() bool {
			return header.PayloadSize <= MaxDataSize && header.UCSize <= MaxDataSize
		}, ErrorHeaderLargePayload},
	}
}

func (header Header) SerializeTo(writer io.Writer) (n int, err error) {
	tempN := 0

	if err = util.Validate(header); err != nil {
		return
	}

	flags := uint16(0)
	if header.IsCompressed() {
		flags |= headerFlagsBitCompressed //leaving this in because why not
	}
	flags |= uint16(header.SignatureType&0b111) << 12

	if err = binary.Write(writer, binary.LittleEndian, flags); err != nil {
		return
	}
	n += binary.Size(flags)

	vw := util.NewVarWriter()
	if tempN, err = vw.WriteUvarint(writer, header.UCSize); err != nil {
		return
	}
	n += tempN
	if tempN, err = vw.WriteUvarint(writer, header.PayloadSize); err != nil {
		return
	}
	n += tempN

	if len(header.Signature) > 0 {
		if tempN, err = writer.Write(header.Signature); err != nil {
			return
		}
		n += tempN
	}

	return
}

//DeserializeFrom tries to deserialize some data from a reader into the given header. The header can be modified on an error return. The header will be validated.
func (header *Header) DeserializeFrom(reader *bufio.Reader) (err error) {
	var flags uint16
	if err = binary.Read(reader, binary.LittleEndian, &flags); err != nil {
		return
	}

	header.SignatureType = itsu_crpyto.SigType((flags & headerFlagsBitsSignatureType) >> 12)

	if header.UCSize, err = binary.ReadUvarint(reader); err != nil {
		return
	}

	if header.PayloadSize, err = binary.ReadUvarint(reader); err != nil {
		return
	}

	sigSize := itsu_crpyto.SignatureSize(header.SignatureType)
	header.Signature = make([]byte, sigSize)

	if _, err = io.ReadFull(reader, header.Signature); err != nil {
		return
	}

	if err = util.Validate(header); err != nil {
		return err
	}

	return
}
