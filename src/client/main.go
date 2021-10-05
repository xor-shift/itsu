package main

import (
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"log"
	"time"
)

func main() {
	var err error
	var tMsg message.Msg

	var session connection.Session
	if session, err = connection.Dial("127.0.0.1:15184"); err != nil {
		log.Panicln(err)
	}

	var handshakeReply message.HandshakeReplyMessage
	if tMsg, _, err = session.WriteAndReadMessageMID(message.HandshakeRequestMessage{SysInfo: util.GetSystemInformation()}, message.MIDHandshakeReply); err != nil {
		log.Panicln(err)
	} else {
		handshakeReply = tMsg.(message.HandshakeReplyMessage)
	}

	log.Println(handshakeReply)

	time.Sleep(time.Millisecond * 100)
}
