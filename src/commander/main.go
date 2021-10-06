package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"fmt"
	g "github.com/AllenDang/giu"
	"log"
	"os"
	"sync"
	"time"
)

var (
	selfID uint64 = 0

	privateKey = ed25519.PrivateKey{71, 220, 40, 69, 141, 59, 87, 127, 121, 248, 224, 195, 161, 44, 104, 59, 32, 217, 62, 144, 11, 154, 181, 168, 79, 67, 42, 195, 179, 57, 209, 172, 251, 50, 163, 155, 192, 130, 254, 58, 208, 73, 2, 244, 16, 223, 215, 128, 223, 112, 174, 97, 211, 46, 48, 76, 59, 2, 146, 26, 12, 143, 221, 97}
)

type State struct {
	session connection.Session
	lastErr error

	token uint64

	serverClientsMutex    *sync.RWMutex
	serverClients         map[uint64]message.ClientInformation
	serverClientsLastSeen map[uint64]time.Time

	threadsWG *sync.WaitGroup
	stopping  chan bool
}

func NewState() (s *State) {
	s = &State{
		session:               connection.Session{},
		lastErr:               nil,
		token:                 0,
		serverClientsMutex:    &sync.RWMutex{},
		serverClients:         make(map[uint64]message.ClientInformation),
		serverClientsLastSeen: make(map[uint64]time.Time),
		threadsWG:             &sync.WaitGroup{},
		stopping:              make(chan bool),
	}

	s.threadsWG.Add(1)
	go s.serverClientsWorker()

	return s
}

func (s *State) serverClientsWorker() {
	ticker := time.NewTicker(time.Second)
	running := true

	for running {
		select {
		case _, _ = <-ticker.C:
			break
		case _, ok := <-s.stopping:
			if !ok {
				running = false
			}
		}
	}
}

func (s *State) refreshServerClients() {

}

func (s *State) Dial(addr string) (err error) {
	s.session, err = connection.Dial(addr)
	s.lastErr = err
	return err
}

func (s *State) GetError() error { return s.lastErr }

func (s *State) RefreshToken() error {
	if reply, _, err := s.session.WriteAndReadMessageMID(message.TokenRequestMessage{}, message.MIDTokenReply); err != nil {
		s.lastErr = err
		return err
	} else {
		s.token = reply.(message.TokenReplyMessage).Token
		return nil
	}
}

func (s *State) GetToken() uint64 { return s.token }

func getClientsRows(list map[uint64]message.ClientInformation) []*g.TableRowWidget {
	rows := make([]*g.TableRowWidget, len(list))

	i := 0
	for k, v := range list {
		rows[i] = g.TableRow(
			g.Label(fmt.Sprint(k)),
			g.Label(fmt.Sprint(v.SysInfo.GONumCPU)))
		i++
	}

	return rows
}

func loop() {
	g.SingleWindow().Layout(
		g.SplitLayout(g.DirectionHorizontal, true, 320,
			g.Layout{
				g.Label("Left pane"),
			},
			g.Layout{
				g.Label("Right pane"),
			}))
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	//window := g.NewMasterWindow("Hello world", 1280, 720, 0)
	//window.Run(loop)

	if len(os.Args) != 1 {
		if os.Args[1] == "genKeys" {
			pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				log.Panicln(err)
			}

			printBytes := func(preamble string, arr []byte) {
				fmt.Print(preamble, "{")

				for k, v := range arr {
					fmt.Print(v)
					if k != (len(arr) - 1) {
						fmt.Print(", ")
					}
				}

				fmt.Print("}\n")
			}

			printBytes("ed25519.PublicKey", pubKey)
			printBytes("ed25519.PrivateKey", privKey)
		}

		os.Exit(0)
	}

	var err error

	var session connection.Session
	if session, err = connection.Dial("127.0.0.1:15184"); err != nil {
		log.Panicln(err)
	}

	if hReply, _, err := session.WriteAndReadMessageMID(message.HandshakeRequestMessage{SysInfo: util.GetSystemInformation()}, message.MIDHandshakeReply); err != nil {
		log.Panicln(err)
	} else {
		selfID = hReply.(message.HandshakeReplyMessage).ID
	}

	log.Println("Established a connection and got an id:", selfID)

	for {
		var clientsList []uint64
		if spr, _, err := session.WriteAndReadMessageED25519MID(&message.ClientsRequestMessage{}, privateKey, message.MIDClientsReply); err != nil {
			log.Panicln(err)
		} else {
			clientsList = spr.(message.ClientsReplyMessage).Clients
		}

		clients := make(map[uint64]message.ClientInformation)
		for _, v := range clientsList {
			if reply, _, err := session.WriteAndReadMessageED25519MID(&message.ClientQueryRequest{ID: v}, privateKey, message.MIDClientQueryReply); err != nil {
				log.Panicln(err)
			} else {
				r := reply.(message.ClientQueryReply)
				if r.Found {
					clients[v] = r.Info
				}
			}
		}

		log.Println(clients)

		time.Sleep(time.Second * 10)
	}
}
