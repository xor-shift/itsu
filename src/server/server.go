package main

import (
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/packet"
	"example.com/itsuMain/lib/util"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Server struct {
	clientsMutex *sync.RWMutex
	clients      map[uint64]*Client

	threadsWG *sync.WaitGroup

	proxyListMutex *sync.RWMutex
	proxyList      []message.ProxyRequest
}

func NewServer() (s *Server) {
	s = &Server{
		clientsMutex: &sync.RWMutex{},
		clients:      make(map[uint64]*Client),

		threadsWG: &sync.WaitGroup{},

		proxyListMutex: &sync.RWMutex{},
		proxyList:      make([]message.ProxyRequest, 0),
	}

	s.threadsWG.Add(1)
	go s.garbageCollector()

	return s
}

func (s *Server) garbageCollector() {
	ticker := time.NewTicker(time.Second * 5)

	clientCollector := func() {
		ids := make([]uint64, 0)

		s.clientsMutex.RLock()
		for k, v := range s.clients {
			if v.done {
				ids = append(ids, k)
			}
		}
		s.clientsMutex.RUnlock()

		if len(ids) > 0 {
			log.Println("Garbage collecting", len(ids), "clients")

			s.clientsMutex.Lock()
			for _, v := range ids {
				delete(s.clients, v)
			}
			s.clientsMutex.Unlock()
		}
	}

	proxyCollector := func() {
		now := time.Now().UnixMilli()

		s.proxyListMutex.Lock()
		defer s.proxyListMutex.Unlock()

		expiredIndices := make([]int, 0)
		for k, v := range s.proxyList {
			if v.ExpiresOn <= now {
				expiredIndices = append(expiredIndices, k)
			}
		}

		util.SliceReverse(expiredIndices)
		for _, v := range expiredIndices {
			util.SliceRemove(&s.proxyList, v)
		}
	}

	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				break
			}
			clientCollector()
			proxyCollector()
		}
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

		threadsWG: &sync.WaitGroup{},
		done:      false,
	}
	s.clients[identifier] = c

	return
}

func (s *Server) deleteClientByID(id uint64) bool {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if c, ok := s.clients[id]; !ok {
		return false
	} else {
		c.Session.Close()
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

	c.threadsWG.Add(1)
	go c.Main(s)

	return
}

func (s *Server) GetClientsList() (list []uint64) {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	list = make([]uint64, len(s.clients))
	i := 0
	for k, _ := range s.clients {
		list[i] = k
		i++
	}

	return
}

func (s *Server) GetClient(id uint64) (c *Client) {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	c, ok := s.clients[id]
	if !ok {
		c = nil
	}

	return
}

func (s *Server) IssueProxyRequest(request message.ProxyRequest) {
	s.proxyListMutex.Lock()
	defer s.proxyListMutex.Unlock()

	s.proxyList = append(s.proxyList, request)
}

func (s *Server) GetProxyRequests(from int64, to int64, cl *Client) []packet.Packet {
	valids := make([]packet.Packet, 0)

	if from <= 0 {
		from = 0
	}

	if to <= 0 {
		to = time.Now().UnixMilli()
	}

	s.proxyListMutex.RLock()
	defer s.proxyListMutex.RUnlock()

	for _, v := range s.proxyList {
		if !(v.IssuedOn >= from && v.IssuedOn <= to) {
			continue
		}

		if linked, err := v.ComparisonProgram.Link(map[string]interface{}{
			"const0": 1.5,
			"const1": "asdasd asd",
		}); err != nil {
			log.Println(err)
			continue
		} else {
			log.Println(linked)
		}

		valids = append(valids, v.Packet)
	}

	return valids
}
