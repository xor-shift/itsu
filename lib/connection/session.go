package connection

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	itsu_crypto "example.com/itsuMain/lib/crpyto"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/packet"
	"github.com/lucas-clemente/quic-go"
	"time"
)

const (
	streamTimeout = time.Second * 3
)

var (
	ErrorUnexpectedMID = errors.New("unexpected MID")
)

type Session struct {
	session quic.Session
	stream  quic.Stream

	reader *bufio.Reader
	writer *bufio.Writer
}

type Listener quic.Listener

func sessionFromQUIC(qSess quic.Session, isClient bool) (session Session, err error) {
	session = Session{
		session: qSess,
		stream:  nil,

		reader: nil,
		writer: nil,
	}

	streamContext, cancel := context.WithTimeout(context.Background(), streamTimeout)
	defer cancel()

	if isClient {
		session.stream, err = qSess.OpenStreamSync(streamContext)
	} else {
		session.stream, err = qSess.AcceptStream(streamContext)
	}

	if err != nil {
		return
	}

	session.reader = bufio.NewReader(session.stream)
	session.writer = bufio.NewWriter(session.stream)

	return
}

func Dial(addr string) (s Session, err error) {
	tlsConf := &tls.Config{
		InsecureSkipVerify:    true,
		NextProtos:            []string{"itsu-comm-proto"},
		VerifyPeerCertificate: itsu_crypto.VerifyServerCertificate,
	}

	quicConf := &quic.Config{
		KeepAlive: false,
	}

	var qSess quic.Session
	if qSess, err = quic.DialAddr(addr, tlsConf, quicConf); err != nil {
		return
	}

	return sessionFromQUIC(qSess, true)
}

func NewListener(addr string) (l Listener, err error) {
	_ = &quic.Config{
		HandshakeIdleTimeout:   5,
		MaxIdleTimeout:         5,
		MaxStreamReceiveWindow: packet.MaxDataSize,
		MaxIncomingStreams:     1,
		MaxIncomingUniStreams:  -1,
		EnableDatagrams:        false,
	}

	var tConf *tls.Config
	if tConf, err = itsu_crypto.NewServerTLSConfig(); err != nil {
		return
	}

	if l, err = quic.ListenAddr(addr, tConf, nil); err != nil {
		l = nil
		return
	}

	return
}

func Accept(listener Listener, ctx context.Context) (s Session, err error) {
	var qSess quic.Session
	if qSess, err = listener.Accept(ctx); err != nil {
		return
	}

	return sessionFromQUIC(qSess, false)
}

func (s *Session) WritePacket(p packet.Packet) (n int, err error) {
	n, err = p.SerializeTo(s.writer)
	if err == nil {
		err = s.writer.Flush()
	}
	return
}

func (s *Session) ReadPacket() (p packet.Packet, err error) {
	err = p.DeserializeFrom(s.reader)
	return
}

func (s *Session) WriteMessage(m message.Msg) (n int, err error) {
	return s.WritePacket(packet.NewPacket(message.SerializeMessage(m)))
}

func (s *Session) ReadMessage() (m message.Msg, p packet.Packet, err error) {
	if p, err = s.ReadPacket(); err != nil {
		return
	}

	if m, err = message.DeserializeMessage(p.Data); err != nil {
		return
	}

	return
}

func (s *Session) ReadMessageMID(id message.MessageID) (m message.Msg, p packet.Packet, err error) {
	if m, p, err = s.ReadMessage(); err != nil {
		return
	}

	if m.GetID() != id {
		err = ErrorUnexpectedMID
		return
	}

	return
}

func (s *Session) WriteAndReadPacket(pOut packet.Packet) (pIn packet.Packet, err error) {
	if _, err = s.WritePacket(pOut); err != nil {
		return
	}

	return s.ReadPacket()
}

func (s *Session) WriteAndReadMessage(mOut message.Msg) (mIn message.Msg, p packet.Packet, err error) {
	if p, err = s.WriteAndReadPacket(packet.NewPacket(message.SerializeMessage(mOut))); err != nil {
		return
	}

	mIn, err = message.DeserializeMessage(p.Data)
	return
}

func (s *Session) WriteAndReadMessageMID(mOut message.Msg, id message.MessageID) (mIn message.Msg, p packet.Packet, err error) {
	if mIn, p, err = s.WriteAndReadMessage(mOut); err != nil {
		return
	}

	if mIn.GetID() != id {
		err = ErrorUnexpectedMID
		return
	}

	return
}
