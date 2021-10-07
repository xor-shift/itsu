package message

import (
	"encoding/gob"
	"example.com/itsuMain/lib/packet"
	"example.com/itsuMain/lib/util"
	"net"
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
	Address string
}

type ClientQueryReply struct {
	Found bool
	Info  ClientInformation
}

func (m ClientQueryReply) GetID() MessageID { return MIDClientQueryReply }

//for integer comparisons:
//-2: the client value must be less than the specified value
//-1: the client value must be less than or equal to the specified value
//0: the client value must be equal to the specified value
//1: the client value must be greater than or equal to the specified value
//2: the client value must be greater than the specified value
//3: the clinet value must not be equal to the specified value
//4: true
//
//for string comparisons:
//true: the client value must equal the specified value
//false: the client value must not equal the specified value
//if the specified string is empty, the result is always true
type ProxyCondition struct {
	RTCPU     int
	RTCPUComp int8

	GOOS     string
	GOOSComp bool

	CPUIDCPU     int
	CPUIDCPUComp int8

	CPUIDHasFeatures         uint64
	CPUIDHasExtendedFeatures uint64
	CPUIDHasExtraFeatures    uint64

	Hostname     string
	HostnameComp bool
	Username     string
	UsernameComp bool
	Address      string
	AddressComp  bool
}

func (c ProxyCondition) CompareWith(information util.SystemInformation, address net.Addr) bool {
	compareInt := func(sv, cv int, typ int8) bool {
		comps := map[int8]func(sv, cv int) bool{
			int8(-2): func(sv, cv int) bool { return cv < sv },
			int8(-1): func(sv, cv int) bool { return cv <= sv },
			int8(0):  func(sv, cv int) bool { return cv == sv },
			int8(1):  func(sv, cv int) bool { return cv >= sv },
			int8(2):  func(sv, cv int) bool { return cv > sv },
			int8(3):  func(sv, cv int) bool { return cv != sv },
			int8(4):  func(sv, cv int) bool { return true },
		}

		if cf, ok := comps[typ]; !ok {
			return false
		} else {
			return cf(sv, cv) || sv == 0
		}
	}

	compareString := func(sv, cv string, typ bool) bool { return typ == (sv == cv) || cv == "" }

	compareMask := func(sv, cv uint64) bool {
		pass := true

		for i := uint64(0); i < 64; i++ {
			pass = pass && ((((sv >> i) & 1) != 1) || (((cv >> i) & 1) == 1)) //p'+q
		}

		return pass
	}

	return compareInt(c.RTCPU, information.GONumCPU, c.RTCPUComp) &&
		compareInt(c.CPUIDCPU, int(information.ProcMaxID), c.CPUIDCPUComp) &&
		compareString(c.GOOS, information.GOOS, c.GOOSComp) &&
		compareMask(c.CPUIDHasFeatures, information.ProcFeatures) &&
		compareMask(c.CPUIDHasExtendedFeatures, information.ProcExtendedFeatures) &&
		compareMask(c.CPUIDHasExtraFeatures, information.ProcExtraFeatures) &&
		compareString(c.Hostname, information.Hostname, c.HostnameComp) &&
		compareString(c.Username, information.Username, c.UsernameComp) &&
		compareString(c.Address, address.String(), c.AddressComp)
}

type ProxyRequest struct {
	MaxTargets int //negative numbers and zero mean broadcast, 1 means regular anycast, any other positive integer means a mix of multi and anycast
	Condition  ProxyCondition

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
