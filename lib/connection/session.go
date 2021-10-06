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
	"net"
	"sync"
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

	tokenLock *sync.Mutex
}

type Listener quic.Listener

func sessionFromQUIC(qSess quic.Session, isClient bool) (session Session, err error) {
	session = Session{
		session: qSess,
		stream:  nil,

		reader: nil,
		writer: nil,

		tokenLock: &sync.Mutex{},
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

func (s *Session) getToken() (uint64, error) {
	if reply, _, err := s.WriteAndReadMessageMID(message.TokenRequestMessage{}, message.MIDTokenReply); err != nil {
		return 0, err
	} else {
		token := reply.(message.TokenReplyMessage).Token
		return token, nil
	}
}

func (s *Session) Address() net.Addr {
	return s.session.RemoteAddr()
}
