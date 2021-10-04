package message

import (
	"encoding/gob"
	"example.com/itsuMain/lib/util"
)

type PingRequestMessage struct{ Token int32 }

func (m PingRequestMessage) GetID() MessageID { return MIDPingRequest }

type PingReplyMessage struct{ Token int32 }

func (m PingReplyMessage) GetID() MessageID { return MIDPingReply }

type HandshakeRequestMessage struct{ SysInfo util.SystemInformation }

func (m HandshakeRequestMessage) GetID() MessageID { return MIDHandshakeRequest }

type HandshakeReplyMessage struct{ ID uint64 }

func (m HandshakeReplyMessage) GetID() MessageID { return MIDHandshakeReply }

type TokenRequestMessage struct{}

func (m TokenRequestMessage) GetID() MessageID { return MIDTokenRequest }

type TokenReplyMessage struct{ Token uint64 }

func (m TokenReplyMessage) GetID() MessageID { return MIDTokenReply }

type SignedPingRequestMessage struct {
	PToken int32
	SToken uint64
}

func (m SignedPingRequestMessage) GetID() MessageID { return MIDSignedPingRequest }

func (m SignedPingRequestMessage) GetSignatureToken() uint64 { return m.SToken }

func (m *SignedPingRequestMessage) SetSignatureToken(v uint64) { m.SToken = v }

type SignedPingReplyMessage struct{ Token int32 }

func (m SignedPingReplyMessage) GetID() MessageID { return MIDSignedPingReply }

func init() {
	gob.Register(PingRequestMessage{})
	gob.Register(PingReplyMessage{})
	gob.Register(PingRequestMessage{})
	gob.Register(PingReplyMessage{})
	gob.Register(HandshakeRequestMessage{})
	gob.Register(HandshakeReplyMessage{})
	gob.Register(TokenRequestMessage{})
	gob.Register(TokenReplyMessage{})
	gob.Register(SignedPingRequestMessage{})
	gob.Register(SignedPingReplyMessage{})
}
