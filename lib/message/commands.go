package message

import "encoding/gob"

type CommandEcho struct {
	Message string
}

func (c CommandEcho) GetID() MessageID { return MIDCmdEcho }

type CommandPanic struct {
	Message string
}

func (c CommandPanic) GetID() MessageID { return MIDCmdPanic }

type CommandShell struct{}

func init() {
	gob.Register(CommandEcho{})
	gob.Register(CommandPanic{})
}
