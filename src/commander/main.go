package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/packet"
	"fmt"
	g "github.com/AllenDang/giu"
	"log"
	"os"
	"time"
)

const (
	fontSize = 12
)

var (
	privateKey = ed25519.PrivateKey{71, 220, 40, 69, 141, 59, 87, 127, 121, 248, 224, 195, 161, 44, 104, 59, 32, 217, 62, 144, 11, 154, 181, 168, 79, 67, 42, 195, 179, 57, 209, 172, 251, 50, 163, 155, 192, 130, 254, 58, 208, 73, 2, 244, 16, 223, 215, 128, 223, 112, 174, 97, 211, 46, 48, 76, 59, 2, 146, 26, 12, 143, 221, 97}
	state      = NewState()

	//ui state

	selectedID = uint64(0)

	CondRTCPU    int32
	CondCPUIDCPU int32
	CondGOOS     int32
	CondAddr     int32

	proxyConditions message.ProxyCondition
	toProxy         message.Msg

	//command stuff
	CmdDuration int32 = 5

	CmdArgsEchoMessage  string
	CmdArgsPanicMessage string
)

func issueCommand(msg message.Msg) {
	proxyConditions.Comparisons[0] = int8(CondRTCPU)
	proxyConditions.Comparisons[1] = int8(CondCPUIDCPU)
	proxyConditions.Comparisons[2] = int8(CondGOOS)

	if _, err := state.session.WriteMessageED25519(&message.ProxyRequest{
		IssuedOn:          time.Now().UnixMilli(),
		ExpiresOn:         time.Now().UnixMilli() + int64(CmdDuration)*1000,
		Packet:            packet.NewPacket(message.SerializeMessage(msg)),
		ComparisonProgram: builtProgram,
	}, privateKey); err != nil {
		log.Panicln(err)
	}
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

	if err := state.Dial("127.0.0.1:15184"); err != nil {
		log.Panicln(err)
	}

	window := g.NewMasterWindow(fmt.Sprint("Given ID: ", state.id), 1280, 720, g.MasterWindowFlagsNotResizable)
	g.SetDefaultFont("FiraCode-Medium", fontSize)
	window.Run(loop)
}
