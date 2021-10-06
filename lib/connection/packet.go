package connection

import (
	"crypto/ed25519"
	itsu_crypto "example.com/itsuMain/lib/crpyto"
	"example.com/itsuMain/lib/packet"
)

func (s *Session) WritePacket(p packet.Packet) (n int, err error) {
	n, err = p.SerializeTo(s.writer)
	if err == nil {
		err = s.writer.Flush()
	}
	return
}

func (s *Session) WritePacketPresigned(p packet.Packet, st itsu_crypto.SigType, signature []byte) (n int, err error) {
	if err = p.PreSign(st, signature); err != nil {
		return
	}

	return s.WritePacket(p)
}

func (s *Session) WritePacketED25519(p packet.Packet, pk ed25519.PrivateKey) (n int, err error) {
	if err = p.SignED25519(pk); err != nil {
		return
	}

	return s.WritePacket(p)
}

func (s *Session) ReadPacket() (p packet.Packet, err error) {
	err = p.DeserializeFrom(s.reader)
	return
}

func (s *Session) WriteAndReadPacket(pOut packet.Packet) (pIn packet.Packet, err error) {
	if _, err = s.WritePacket(pOut); err != nil {
		return
	}

	return s.ReadPacket()
}

func (s *Session) WriteAndReadPacketPresigned(pOut packet.Packet, st itsu_crypto.SigType, signature []byte) (pIn packet.Packet, err error) {
	if _, err = s.WritePacketPresigned(pOut, st, signature); err != nil {
		return
	}

	return s.ReadPacket()
}

func (s *Session) WriteAndReadPacketED25519(pOut packet.Packet, pk ed25519.PrivateKey) (pIn packet.Packet, err error) {
	if _, err = s.WritePacketED25519(pOut, pk); err != nil {
		return
	}

	return s.ReadPacket()
}
