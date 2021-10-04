package main

import (
	"context"
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"log"
	"math/rand"
	"sync"
)

type Server struct {
	clientsMutex *sync.RWMutex
	clients      map[uint64]*Client
}

func NewServer() (s *Server) {
	return &Server{
		clientsMutex: &sync.RWMutex{},
		clients:      make(map[uint64]*Client),
	}
}

func (s *Server) allocateNewClient() (c *Client, identifier uint64) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	identifier = rand.Uint64()
	for {
		if _, ok := s.clients[identifier]; !ok {
			break
		}
		identifier = rand.Uint64()
	}

	c = &Client{
		Session: connection.Session{},

		identifier: 0,
		sysInfo:    util.SystemInformation{},

		currentToken: 0,
	}
	s.clients[identifier] = c

	return
}

func (s *Server) deleteClientByID(id uint64) bool {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if _, ok := s.clients[id]; !ok {
		return false
	}

	delete(s.clients, id)

	return true
}

func (s *Server) NewClient(sess connection.Session) (c *Client, err error) {
	var handshakeRequest message.HandshakeRequestMessage
	if tempMessage, _, err := sess.ReadMessageMID(message.MIDHandshakeRequest); err != nil {
		return nil, err
	} else {
		handshakeRequest = tempMessage.(message.HandshakeRequestMessage)
	}

	var identifier uint64
	c, identifier = s.allocateNewClient()

	c.Session = sess
	c.identifier = identifier
	c.sysInfo = handshakeRequest.SysInfo
	c.currentToken = rand.Uint64()

	if _, err = c.Session.WriteMessage(message.HandshakeReplyMessage{ID: identifier}); err != nil {
		s.deleteClientByID(identifier)
		return
	}

	return
}

type Client struct {
	Session connection.Session

	identifier uint64
	sysInfo    util.SystemInformation

	currentToken uint64
}

func main() {
	var err error
	var listener connection.Listener

	server := NewServer()

	if listener, err = connection.NewListener("0.0.0.0:15184"); err != nil {
		log.Panicln(err)
	}

	for {
		var s connection.Session
		if s, err = connection.Accept(listener, context.Background()); err != nil {
			log.Println(err)
			continue
		}

		if c, err := server.NewClient(s); err != nil {
			log.Println("Error while accepting a new client:", err)
		} else {
			log.Println("Accepted new client with ID", c.identifier)
		}
	}
}
