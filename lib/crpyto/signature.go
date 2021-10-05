package itsu_crypto

import (
	"crypto/ed25519"
	"errors"
)

type SigType uint

const (
	SigTypeNone    = 0
	SigTypeED25519 = 1
	SigTypeMax     = 7
)

var (
	trustedClientKeys = map[SigType][]interface{}{
		SigTypeNone: {},
		SigTypeED25519: {
			ed25519.PublicKey{251, 50, 163, 155, 192, 130, 254, 58, 208, 73, 2, 244, 16, 223, 215, 128, 223, 112, 174, 97, 211, 46, 48, 76, 59, 2, 146, 26, 12, 143, 221, 97}},
	}

	ErrorClientSigInternal = errors.New("internal error while verifying client signature")
	ErrorClientSigBadToken = errors.New("invalid or expired signature token")
	ErrorClientSigUnsigned = errors.New("message requiring signature is unsigned")
)

func SignatureSize(sType SigType) int {
	sizeMap := map[SigType]int{
		SigTypeNone:    0,
		SigTypeED25519: ed25519.SignatureSize,
	}

	if v, ok := sizeMap[sType]; ok {
		return v
	} else {
		return 0
	}
}

func VerifyClientSignature(data []byte, signature []byte, sigType SigType) (err error) {
	/*if !message.GetMIDProperties(msg.GetID()).RequiresSignature {
		return nil
	}

	if signedMsg, ok := msg.(message.SignedMessage); !ok {
		return ErrorClientSigInternal
	} else {
		if expectedToken != signedMsg.GetSignatureToken() {
			return ErrorClientSigBadToken
		}
	}*/

	if expectedSize := SignatureSize(sigType); (expectedSize == 0) || (expectedSize != len(signature)) {
		return ErrorClientSigUnsigned
	}

	ok := false
	switch sigType {
	case 0:
		break
	case 1:
		for _, v := range trustedClientKeys[SigTypeED25519] {
			ok = ok || ed25519.Verify(v.(ed25519.PublicKey), data, signature)
		}
	default:
		break
	}

	if !ok {
		return ErrorClientSigUnsigned
	}

	return nil
}
