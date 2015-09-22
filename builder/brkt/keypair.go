package brkt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
)

type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

func NewRandomKeyPair() (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyDer,
	}
	privateKeyPem := string(pem.EncodeToMemory(&privateKeyBlock))

	publicKey := privateKey.PublicKey
	pub, err := ssh.NewPublicKey(&publicKey)
	if err != nil {
		return nil, err
	}
	pubBytes := ssh.MarshalAuthorizedKey(pub)

	return &KeyPair{
		PublicKey:  string(pubBytes),
		PrivateKey: privateKeyPem,
	}, nil
}
