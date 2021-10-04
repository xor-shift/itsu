package message

import "errors"

type MessageID uint32

const (
	/*
		bit layout:
		rccc iiii iiii
		r -> reply?
		c -> category
		i -> id
	*/

	midReplyBit = 0x800
	midCat0     = 0x000 // misc.
	midCat1     = 0x100 // state
	midCat2     = 0x200 // server query
	midCat3     = 0x300 // server management
	midCat4     = 0x400 // c&c c->s
	midCat5     = 0x500 // c&c s->c
	midCat6     = 0x600 // reserved
	midCat7     = 0x700 // general errors and statuses

	//

	MIDPingRequest       = midCat0 | 0
	MIDSignedPingRequest = midCat0 | 1

	MIDHandshakeRequest = midCat1 | 0
	MIDTokenRequest     = midCat1 | 1

	//

	MIDErrorBadRequest = midReplyBit | midCat7 | 0
	MIDErrorInternal   = midReplyBit | midCat7 | 1
	MIDErrorUnsigned   = midReplyBit | midCat7 | 2

	//

	MIDPingReply       = midReplyBit | MIDPingRequest
	MIDSignedPingReply = midReplyBit | MIDSignedPingRequest

	MIDHandshakeReply = midReplyBit | MIDHandshakeRequest
	MIDTokenReply     = midReplyBit | MIDTokenRequest

	//

	MIDInvalid = MessageID(0xFFF)
)

type MIDProperties struct {
	RequiresSignature bool
}

var (
	ErrorUnexpectedMID = errors.New("unexpected MID")
	ErrorUnknownMID    = errors.New("unknown MID")

	MIDPropertyMap = map[MessageID]MIDProperties{
		MIDSignedPingRequest: {true},
	}
)

func GetMIDProperties(mid MessageID) MIDProperties {
	if p, ok := MIDPropertyMap[mid]; ok {
		return p
	} else {
		return MIDProperties{
			RequiresSignature: false,
		}
	}
}
