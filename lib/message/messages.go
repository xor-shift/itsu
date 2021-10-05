package message

import (
	"encoding/gob"
	"example.com/itsuMain/lib/packet"
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

type ClientsRequestMessage struct{ Token uint64 }

func (m ClientsRequestMessage) GetID() MessageID { return MIDClientsRequest }

func (m ClientsRequestMessage) GetSignatureToken() uint64 { return m.Token }

func (m *ClientsRequestMessage) SetSignatureToken(v uint64) { m.Token = v }

type ClientsReplyMessage struct{ Clients []uint64 }

func (m ClientsReplyMessage) GetID() MessageID { return MIDClientsReply }

type ClientQueryRequest struct {
	Token uint64
	ID    uint64
}

func (m ClientQueryRequest) GetID() MessageID { return MIDClientQueryRequest }

func (m ClientQueryRequest) GetSignatureToken() uint64 { return m.Token }

func (m *ClientQueryRequest) SetSignatureToken(v uint64) { m.Token = v }

type ClientInformation struct {
	SysInfo util.SystemInformation
}

type ClientQueryReply struct {
	Found bool
	Info  ClientInformation
}

func (m ClientQueryReply) GetID() MessageID { return MIDClientQueryReply }

type ProxyRequest struct {
	/*
		0 -> unicast
		1 -> broadcast
		2 -> multicast
		3 -> anycast
	*/
	Type int

	//for type 0 and 1
	Target uint64

	//for type 2 and 3
	Targets []uint64

	//for type 3
	MaxRelays int

	Packet packet.Packet

	Token uint64
}

func (m ProxyRequest) GetID() MessageID { return MIDProxyRequestMessage }

func (m ProxyRequest) GetSignatureToken() uint64 { return m.Token }

func (m *ProxyRequest) SetSignatureToken(v uint64) { m.Token = v }

type ProxyReply struct {
	RelayedTo []uint64
}

func (m ProxyReply) GetID() MessageID { return MIDProxyReplyMessage }

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
	gob.Register(ClientsRequestMessage{})
	gob.Register(ClientsReplyMessage{})
	gob.Register(ClientQueryRequest{})
	gob.Register(ClientQueryReply{})
	gob.Register(ProxyRequest{})
	gob.Register(ProxyReply{})
}
