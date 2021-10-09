package main

import (
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"log"
	"time"
)

const (
	connectionPeriod = time.Second * 6
)

var (
	lastFetch int64 = 0
)

func main() {
	for {
		mainFunc()
		time.Sleep(connectionPeriod)
	}
}

func mainFunc() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Client recovered from panic:", r)
		}
	}()

	var err error
	var tMsg message.Msg

	var session connection.Session
	if session, err = connection.Dial("127.0.0.1:15184"); err != nil {
		log.Panicln(err)
	}

	defer func() {
		if e0, e1 := session.Close(); e0 != nil || e1 != nil {
			log.Print("Session closure errored: ", e0, ", ", e1)
		}
	}()

	var clientID uint64
	if tMsg, _, err = session.WriteAndReadMessageMID(message.HandshakeRequestMessage{SysInfo: util.GetSystemInformation()}, message.MIDHandshakeReply); err != nil {
		log.Panicln(err)
	} else {
		clientID = tMsg.(message.HandshakeReplyMessage).ID
	}
	_ = clientID

	//established a connection, query for commands now
	if _, err = session.WriteMessage(message.FetchProxyRequest{From: lastFetch}); err != nil {
		log.Panicln(err)
	}
	lastFetch = time.Now().UnixMilli()

	for lastMsg := message.Msg(nil); ; {
		if lastMsg, _, err = session.ReadMessage(); err != nil {
			log.Panicln(err)
		}

		if lastMsg.GetID() == message.MIDFetchProxyReply {
			break
		}

		switch v := lastMsg.(type) {
		case message.CommandEcho:
			log.Println(v.Message)
			break
		case message.CommandPanic:
			panic(v.Message)
		default:
			log.Println("Unhandled MID for commands:", lastMsg.GetID())
		}
	}
}
