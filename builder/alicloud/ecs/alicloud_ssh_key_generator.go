package ecs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	gossh "golang.org/x/crypto/ssh"
)

var (
	ErrKeyGeneration = errors.New("Unable to generate key")
	ErrValidation    = errors.New("Unable to validate key")
	ErrPublicKey     = errors.New("Unable to convert public key")
)

type KeyPair struct {
	PrivateKey string
	PublicKey  string
}

// NewKeyPair generates a new SSH keypair
// This will return a private & public key encoded as DER.
func NewKeyPair() (keyPair *KeyPair, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, ErrKeyGeneration
	}

	if err := priv.Validate(); err != nil {
		return nil, ErrValidation
	}

	privDer := x509.MarshalPKCS1PrivateKey(priv)

	pubSSH, err := gossh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, ErrPublicKey
	}

	return &KeyPair{
		PrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Headers: nil, Bytes: privDer})),
		PublicKey:  string(gossh.MarshalAuthorizedKey(pubSSH)),
	}, nil
}
