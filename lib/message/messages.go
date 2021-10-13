package message

import (
	"encoding/gob"
	"example.com/itsuMain/lib/packet"
	"example.com/itsuMain/lib/util"
	"example.com/itsuMain/lib/vm"
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

// ProxyCondition
/*asd
comparisons:
-3: cv < sv
-2: cv <= sv
-1: cv != sv
0 : true
1 : cv == sv
2 : cv >= sv
3 : cv > sv
*/
type ProxyCondition struct {
	Comparisons [6]int8

	RTCPU    int32
	CPUIDCPU int32
	GOOS     string
	Hostname string
	Username string
	Address  string

	CPUIDFeatures         uint64
	CPUIDExtendedFeatures uint64
	CPUIDExtraFeatures    uint64
}

const (
	CompFieldRTCPU    = 0
	CompFieldCPUIDCPU = 1
	CompFieldGOOS     = 2
	CompFieldHostname = 3
	CompFieldUsername = 4
)

func (c ProxyCondition) CompareWith(information util.SystemInformation, address net.Addr) bool {
	compareInt := func(sv, cv int32, typ int8) bool {
		comps := map[int8]func(sv, cv int32) bool{
			int8(-3): func(sv, cv int32) bool { return cv < sv },
			int8(-2): func(sv, cv int32) bool { return cv <= sv },
			int8(-1): func(sv, cv int32) bool { return cv != sv },
			int8(0):  func(sv, cv int32) bool { return true },
			int8(1):  func(sv, cv int32) bool { return cv == sv },
			int8(2):  func(sv, cv int32) bool { return cv >= sv },
			int8(3):  func(sv, cv int32) bool { return cv > sv },
		}

		if cf, ok := comps[typ]; !ok {
			return false
		} else {
			return cf(sv, cv) || sv == 0
		}
	}

	compareString := func(sv, cv string, typ int8) bool {
		return (typ == 0) || (typ == 1 && sv == cv) || (typ == -1 && sv != cv)
	}

	compareMask := func(sv, cv uint64) bool {
		pass := true

		for i := uint64(0); i < 64; i++ {
			pass = pass && ((((sv >> i) & 1) != 1) || (((cv >> i) & 1) == 1)) //p'+q
		}

		return pass
	}

	return compareInt(c.RTCPU, int32(information.GONumCPU), c.Comparisons[0]) &&
		compareInt(c.CPUIDCPU, int32(information.ProcMaxID), c.Comparisons[1]) &&
		compareString(c.GOOS, information.GOOS, c.Comparisons[2]) &&
		compareString(c.Hostname, information.Hostname, c.Comparisons[3]) &&
		compareString(c.Username, information.Username, c.Comparisons[4]) &&
		compareString(c.Address, address.String(), c.Comparisons[5]) &&
		compareMask(c.CPUIDFeatures, information.ProcFeatures) &&
		compareMask(c.CPUIDExtendedFeatures, information.ProcExtendedFeatures) &&
		compareMask(c.CPUIDExtraFeatures, information.ProcExtraFeatures)
}

type ProxyRequest struct {
	MaxTargets        int //negative numbers and zero mean broadcast, 1 means regular anycast, any other positive integer means a mix of multi and anycast
	ComparisonProgram vm.BuiltProgram

	IssuedOn  int64 //the date at which the proxy is issued
	ExpiresOn int64 //the date at which the proxy expires

	Packet packet.Packet

	Token uint64
}

func (m ProxyRequest) GetID() MessageID { return MIDProxyRequest }

func (m ProxyRequest) GetSignatureToken() uint64 { return m.Token }

func (m *ProxyRequest) SetSignatureToken(v uint64) { m.Token = v }

type ProxyReply struct {
	RelayedTo []uint64
}

func (m ProxyReply) GetID() MessageID { return MIDProxyReply }

type FetchProxyRequest struct {
	From, To int64
}

func (m FetchProxyRequest) GetID() MessageID { return MIDFetchProxyRequest }

//FetchProxyReply is sent at the end of a proxy message stream
type FetchProxyReply struct{}

func (m FetchProxyReply) GetID() MessageID { return MIDFetchProxyReply }

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
	gob.Register(FetchProxyRequest{})
	gob.Register(FetchProxyReply{})
}
