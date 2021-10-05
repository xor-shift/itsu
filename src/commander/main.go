package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	selfID uint64 = 0

	privateKey = ed25519.PrivateKey{71, 220, 40, 69, 141, 59, 87, 127, 121, 248, 224, 195, 161, 44, 104, 59, 32, 217, 62, 144, 11, 154, 181, 168, 79, 67, 42, 195, 179, 57, 209, 172, 251, 50, 163, 155, 192, 130, 254, 58, 208, 73, 2, 244, 16, 223, 215, 128, 223, 112, 174, 97, 211, 46, 48, 76, 59, 2, 146, 26, 12, 143, 221, 97}
)

func requestSigToken(sess connection.Session) (token uint64, err error) {
	if reply, _, err := sess.WriteAndReadMessageMID(message.TokenRequestMessage{}, message.MIDTokenReply); err != nil {
		return 0, err
	} else {
		token = reply.(message.TokenReplyMessage).Token
	}

	return
}

func forceRequestSigToken(sess connection.Session) (token uint64) {
	var err error
	if token, err = requestSigToken(sess); err != nil {
		log.Panicln(err)
	}
	return
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

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
		if spr, _, err := session.WriteAndReadMessageED25519MID(&message.ClientsRequestMessage{}, forceRequestSigToken(session), privateKey, message.MIDClientsReply); err != nil {
			log.Panicln(err)
		} else {
			clientsList = spr.(message.ClientsReplyMessage).Clients
		}

		clients := make(map[uint64]message.ClientInformation)
		for _, v := range clientsList {
			if reply, _, err := session.WriteAndReadMessageED25519MID(&message.ClientQueryRequest{ID: v}, forceRequestSigToken(session), privateKey, message.MIDClientQueryReply); err != nil {
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
