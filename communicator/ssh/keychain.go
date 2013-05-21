package ssh

import (
	"crypto"
	"crypto/dsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

type SimpleKeychain struct {
	keys []interface{}
}

// AddPEMKey adds a simple PEM encoded private key to the keychain.
func (k *SimpleKeychain) AddPEMKey(key string) (err error) {
	block, _ := pem.Decode([]byte(key))
	rsakey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return
	}

	k.keys = append(k.keys, rsakey)
	return
}

// Key method for ssh.ClientKeyring interface
func (k *SimpleKeychain) Key(i int) (interface{}, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}
	switch key := k.keys[i].(type) {
	case *rsa.PrivateKey:
		return &key.PublicKey, nil
	case *dsa.PrivateKey:
		return &key.PublicKey, nil
	}
	panic("unknown key type")
}

// Sign method for ssh.ClientKeyring interface
func (k *SimpleKeychain) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
	hashFunc := crypto.SHA1
	h := hashFunc.New()
	h.Write(data)
	digest := h.Sum(nil)
	switch key := k.keys[i].(type) {
	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand, key, hashFunc, digest)
	}
	return nil, errors.New("ssh: unknown key type")
}
