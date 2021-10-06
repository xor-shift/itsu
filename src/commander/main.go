package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"example.com/itsuMain/lib/util"
	"fmt"
	g "github.com/AllenDang/giu"
	"log"
	"math"
	"os"
	"sort"
	"time"
)

var (
	privateKey = ed25519.PrivateKey{71, 220, 40, 69, 141, 59, 87, 127, 121, 248, 224, 195, 161, 44, 104, 59, 32, 217, 62, 144, 11, 154, 181, 168, 79, 67, 42, 195, 179, 57, 209, 172, 251, 50, 163, 155, 192, 130, 254, 58, 208, 73, 2, 244, 16, 223, 215, 128, 223, 112, 174, 97, 211, 46, 48, 76, 59, 2, 146, 26, 12, 143, 221, 97}
	state      = NewState()
)

func getClientsRows() []*g.TableRowWidget {
	type intermediate struct {
		id       uint64
		lastSeen int
		sysInfo  util.SystemInformation
		addr     string
	}

	intermediates := make([]intermediate, len(state.serverClientsLastSeen))

	state.serverClientsMutex.Lock()

	i := 0
	for k, v := range state.serverClientsLastSeen {
		intermediates[i].id = k
		intermediates[i].lastSeen = (int)(math.Trunc(time.Now().Sub(v).Seconds()))
		intermediates[i].sysInfo = state.serverClients[k].SysInfo
		intermediates[i].addr = state.serverClients[k].Address
		i++
	}

	state.serverClientsMutex.Unlock()

	sort.Slice(intermediates, func(i, j int) bool { return intermediates[i].id < intermediates[j].id })

	rows := make([]*g.TableRowWidget, len(intermediates))

	for k, v := range intermediates {
		rows[k] = g.TableRow(
			g.Label(fmt.Sprint(v.id)),
			g.Label(v.addr),
			g.Label(fmt.Sprint(v.lastSeen)),
		)
	}

	return rows
}

func loop() {
	g.SingleWindow().Layout(
		g.SplitLayout(g.DirectionHorizontal, true, 320,
			g.Layout{
				g.Table().
					Columns(
						g.TableColumn("ID").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(110),
						g.TableColumn("Address").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(80),
						g.TableColumn("Secs").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(25),
					).
					Freeze(0, 1).
					FastMode(true).
					Rows(getClientsRows()...),
			},
			g.Layout{
				g.Label("Right pane"),
			}))
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

	window := g.NewMasterWindow("Hello world", 1280, 720, 0)
	g.SetDefaultFont("FiraCode-Medium", 12)
	window.Run(loop)
}
