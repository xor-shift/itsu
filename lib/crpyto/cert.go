package itsu_crypto

import (
	"crypto/ed25519"
	"crypto/x509"
	"errors"
)

var (
	trustedServerKeys = map[SigType][]interface{}{
		SigTypeNone: {},
		SigTypeED25519: {
			ed25519.PublicKey{57, 65, 106, 56, 54, 140, 122, 209, 148, 68, 14, 215, 170, 163, 181, 74, 211, 22, 159, 110, 67, 12, 248, 246, 221, 62, 223, 103, 216, 64, 144, 17},
		},
	}

	ErrorCertMissing    = errors.New("missing certificate(s)")
	ErrorCertUnverified = errors.New("could not verify certificate(s)")
)

func VerifyServerCertificate(rawCerts [][]byte, _ [][]*x509.Certificate) (err error) {
	if len(rawCerts) == 0 {
		return ErrorCertMissing
	}

	for _, v := range rawCerts {
		if err = verifySingleCert(v); err != nil {
			return nil
		}
	}

	return ErrorCertUnverified
}

func verifySingleCert(rawCert []byte) error {
	var err error
	var cert *x509.Certificate

	if cert, err = x509.ParseCertificate(rawCert); err != nil {
		return err
	}

	switch cert.PublicKey.(type) {
	case ed25519.PublicKey:
		if verifyED25519ServerCert(cert.PublicKey.(ed25519.PublicKey)) {
			return nil
		}
	default:
		break
	}

	return ErrorCertUnverified
}

func verifyED25519ServerCert(pubkey ed25519.PublicKey) bool {
	for _, tKey := range trustedServerKeys[SigTypeED25519] {
		tKeyED := tKey.(ed25519.PublicKey)

		ok := true
		for i := 0; i < ed25519.PublicKeySize; i++ {
			ok = ok && (pubkey[i] == tKeyED[i])
		}
		if ok {
			return true
		}
	}

	return false
}
