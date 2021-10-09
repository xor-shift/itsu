package main

import (
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	staleClientDuration = time.Second * 5
	refreshDuration     = time.Second * 1
)

type State struct {
	session connection.Session
	id      uint64
	lastErr error

	serverClientsMutex    *sync.RWMutex
	serverClients         map[uint64]message.ClientInformation
	serverClientsLastSeen map[uint64]time.Time

	threadsWG *sync.WaitGroup
	stopping  chan bool

	refreshesPaused atomic.Value
}

func NewState() (s *State) {
	s = &State{
		session: connection.Session{},
		lastErr: nil,

		serverClientsMutex:    &sync.RWMutex{},
		serverClients:         make(map[uint64]message.ClientInformation),
		serverClientsLastSeen: make(map[uint64]time.Time),

		threadsWG: &sync.WaitGroup{},
		stopping:  make(chan bool),

		refreshesPaused: atomic.Value{},
	}

	s.refreshesPaused.Store(false)

	return s
}

func (s *State) serverClientsWorker() {
	ticker := time.NewTicker(refreshDuration)

	running := true

	refresh := func() {
		if s.refreshesPaused.Load().(bool) {
			return
		}
		if err := s.refreshServerClients(); err != nil {
			log.Println("client list refresh returned error:", err)
		}
	}

	refresh()

	for running {
		select {
		case _, _ = <-ticker.C:
			refresh()
			break
		case _, ok := <-s.stopping:
			if !ok {
				running = false
			}
		}
	}
}

func (s *State) refreshServerClients() error {
	var clientsList []uint64
	if reply, _, err := s.session.WriteAndReadMessageED25519MID(&message.ClientsRequestMessage{}, privateKey, message.MIDClientsReply); err != nil {
		return err
	} else {
		clientsList = reply.(message.ClientsReplyMessage).Clients
	}

	tempClients := make(map[uint64]message.ClientInformation)
	for _, v := range clientsList {
		if reply, _, err := s.session.WriteAndReadMessageED25519MID(&message.ClientQueryRequest{ID: v}, privateKey, message.MIDClientQueryReply); err != nil {
			return err
		} else {
			r := reply.(message.ClientQueryReply)
			if r.Found {
				tempClients[v] = r.Info
			}
		}
	}

	s.serverClientsMutex.Lock()
	defer s.serverClientsMutex.Unlock()

	for k, v := range tempClients {
		s.serverClientsLastSeen[k] = time.Now()
		s.serverClients[k] = v
	}

	toPrune := make([]uint64, 0)
	pruneTime := time.Now()
	for k, v := range s.serverClientsLastSeen {
		if pruneTime.Sub(v) >= staleClientDuration {
			toPrune = append(toPrune, k)
		}
	}

	if len(toPrune) > 0 {
		log.Println("Pruning", len(toPrune), "stale client entries")

		for _, v := range toPrune {
			delete(s.serverClients, v)
			delete(s.serverClientsLastSeen, v)
		}
	}

	return nil
}

func (s *State) Dial(addr string) (err error) {
	if s.session, err = connection.Dial(addr); err != nil {
		s.lastErr = err
		return
	}

	if reply, _, err := s.session.WriteAndReadMessageMID(message.HandshakeRequestMessage{SysInfo: util.GetSystemInformation()}, message.MIDHandshakeReply); err != nil {
		s.lastErr = err
		return err
	} else {
		s.id = reply.(message.HandshakeReplyMessage).ID
	}

	s.threadsWG.Add(1)
	go s.serverClientsWorker()

	return
}

func (s *State) GetError() error { return s.lastErr }

func (s *State) IsRefreshing() bool {
	return !s.refreshesPaused.Load().(bool)
}

func (s *State) ToggleRefreshes() bool {
	v := s.IsRefreshing()
	s.refreshesPaused.Store(v)
	return v
}
