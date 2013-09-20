package ssh

import (
	"code.google.com/p/go.crypto/ssh"
	"crypto"
	"crypto/dsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

// SimpleKeychain makes it easy to use private keys in order to connect
// via SSH, since the interface exposed by Go isn't the easiest to use
// right away.
type SimpleKeychain struct {
	keys []interface{}
}

// AddPEMKey adds a simple PEM encoded private key to the keychain.
func (k *SimpleKeychain) AddPEMKey(key string) (err error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return errors.New("no block in key")
	}

	var rsakey interface{}
	rsakey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		rsakey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	}

	if err != nil {
		return
	}

	k.keys = append(k.keys, rsakey)
	return
}

// AddPEMKeyPassword adds a PEM encoded private key that is protected by
// a password to the keychain.
func (k *SimpleKeychain) AddPEMKeyPassword(key string, password string) (err error) {
	block, _ := pem.Decode([]byte(key))
	bytes, _ := x509.DecryptPEMBlock(block, []byte(password))
	rsakey, err := x509.ParsePKCS1PrivateKey(bytes)
	if err != nil {
		return
	}

	k.keys = append(k.keys, rsakey)
	return
}

// Key method for ssh.ClientKeyring interface
func (k *SimpleKeychain) Key(i int) (ssh.PublicKey, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}
	switch key := k.keys[i].(type) {
	case *rsa.PrivateKey:
		return ssh.NewPublicKey(&key.PublicKey)
	case *dsa.PrivateKey:
		return ssh.NewPublicKey(&key.PublicKey)
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
