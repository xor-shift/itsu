package itsu_crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
)

var (
	serverPrivkey = ed25519.PrivateKey{8, 202, 19, 225, 13, 89, 65, 101, 7, 223, 178, 92, 214, 222, 61, 239, 255, 234, 170, 91, 41, 223, 171, 234, 1, 218, 94, 108, 52, 21, 95, 143, 136, 37, 110, 186, 157, 144, 165, 71, 234, 250, 238, 24, 175, 69, 35, 227, 255, 185, 94, 56, 88, 23, 132, 21, 140, 2, 49, 97, 148, 246, 237, 111}
	serverPubkey  = ed25519.PublicKey{136, 37, 110, 186, 157, 144, 165, 71, 234, 250, 238, 24, 175, 69, 35, 227, 255, 185, 94, 56, 88, 23, 132, 21, 140, 2, 49, 97, 148, 246, 237, 111}
)

func NewServerTLSConfig() (tConf *tls.Config, err error) {
	template := x509.Certificate{SerialNumber: big.NewInt(1337)}

	var certDER []byte
	if certDER, err = x509.CreateCertificate(rand.Reader, &template, &template, serverPubkey, serverPrivkey); err != nil {
		panic(err)
	}

	var pemKeyBytes []byte
	if pemKeyBytes, err = x509.MarshalPKCS8PrivateKey(serverPrivkey); err != nil {
		panic(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "ED25519 PRIVATE KEY", Bytes: pemKeyBytes})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	var tlsCert tls.Certificate
	if tlsCert, err = tls.X509KeyPair(certPEM, keyPEM); err != nil {
		return
	}

	tConf = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"itsu-comm-proto"},
	}

	return
}
