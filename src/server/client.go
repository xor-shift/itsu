package main

import (
	"errors"
	"example.com/itsuMain/lib/connection"
	itsu_crypto "example.com/itsuMain/lib/crpyto"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/packet"
	"example.com/itsuMain/lib/util"
	"fmt"
	"log"
	"math/rand"
	"sync"
)

type Client struct {
	Session connection.Session

	identifier uint64
	sysInfo    util.SystemInformation

	currentToken uint64

	threadsWG *sync.WaitGroup
	done      bool //for gc, async reads that have race conditions doesn't matter
}

var (
	ErrorUnhandledMID = errors.New("unhandled MID")
)

type clientLogger struct {
	identifier uint64
}

func (c clientLogger) println(v ...interface{}) {
	log.Printf("[%d]: %s", c.identifier, fmt.Sprint(v...))
}

func (c *Client) logger() clientLogger {
	return clientLogger{identifier: c.identifier}
}

func (c *Client) Main(s *Server) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Client worker for id", c.identifier, "recovered:", r)
		}
	}()

	defer func() {
		c.threadsWG.Done()
		c.threadsWG.Wait()
		c.done = true
	}()

	logger := c.logger()

	logger.println("worker started")

	for {
		if msg, p, err := c.Session.ReadMessage(); err != nil {
			logger.println("couldn't read message: ", err)
			break
		} else {
			logger.println("received a message with mid: ", msg.GetID())

			if err = c.handleMessage(s, msg, p); err != nil {
				logger.println("message handler returned error: ", err)
				break
			}
		}
	}

	logger.println("worker stopping")
}

func (c *Client) handleMessage(s *Server, m message.Msg, p packet.Packet) (err error) {
	if message.GetMIDProperties(m.GetID()).RequiresSignature {
		if signedM, ok := m.(message.SignedMessage); !ok {
			return itsu_crypto.ErrorClientSigInternal
		} else {
			if signedM.GetSignatureToken() != c.currentToken {
				err = itsu_crypto.ErrorClientSigBadToken
			} else {
				c.currentToken = rand.Uint64()
			}

			if err = itsu_crypto.VerifyClientSignature(p.Data, p.Signature, p.SignatureType); err != nil {
				return
			}
		}
	}

	switch v := m.(type) {
	case message.PingRequestMessage:
		_, err = c.Session.WriteMessage(message.PingReplyMessage{Token: v.Token})
		break
	case message.SignedPingRequestMessage:
		_, err = c.Session.WriteMessage(message.SignedPingReplyMessage{Token: v.PToken})
		break
	case message.TokenRequestMessage:
		_, err = c.Session.WriteMessage(message.TokenReplyMessage{Token: c.currentToken})
		if err != nil {
			c.currentToken = rand.Uint64()
		}
		break
	case message.ClientsRequestMessage:
		list := s.GetClientsList()
		_, err = c.Session.WriteMessage(message.ClientsReplyMessage{Clients: list})
	case message.ClientQueryRequest:
		cl := s.GetClient(v.ID)
		reply := message.ClientQueryReply{}

		if cl == nil {
			reply.Found = false
		} else {
			reply.Found = true
			reply.Info.SysInfo = cl.sysInfo
		}

		_, err = c.Session.WriteMessage(reply)
	default:
		c.logger().println("unhandled MID: ", m.GetID())
		err = ErrorUnhandledMID
		break
	}

	return
}
