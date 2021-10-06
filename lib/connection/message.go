package connection

import (
	"crypto/ed25519"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/packet"
)

func (s *Session) WriteMessage(m message.Msg) (n int, err error) {
	return s.WritePacket(packet.NewPacket(message.SerializeMessage(m)))
}

func (s *Session) WriteMessageED25519(m message.SignableMessage, pk ed25519.PrivateKey) (n int, err error) {
	s.tokenLock.Lock()
	defer s.tokenLock.Unlock()

	var sigToken uint64
	if sigToken, err = s.getToken(); err != nil {
		return
	}

	m.SetSignatureToken(sigToken)
	return s.WritePacketED25519(packet.NewPacket(message.SerializeMessage(m)), pk)
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

func (s *Session) WriteAndReadMessage(mOut message.Msg) (mIn message.Msg, p packet.Packet, err error) {
	if _, err = s.WriteMessage(mOut); err != nil {
		return
	}

	mIn, p, err = s.ReadMessage()
	return
}

func (s *Session) WriteAndReadMessageED25519(mOut message.SignableMessage, pk ed25519.PrivateKey) (mIn message.Msg, p packet.Packet, err error) {
	if _, err = s.WriteMessageED25519(mOut, pk); err != nil {
		return
	}

	mIn, p, err = s.ReadMessage()
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

func (s *Session) WriteAndReadMessageED25519MID(mOut message.SignableMessage, pk ed25519.PrivateKey, id message.MessageID) (mIn message.Msg, p packet.Packet, err error) {
	if mIn, p, err = s.WriteAndReadMessageED25519(mOut, pk); err != nil {
		return
	}

	if mIn.GetID() != id {
		err = ErrorUnexpectedMID
		return
	}

	return
}
