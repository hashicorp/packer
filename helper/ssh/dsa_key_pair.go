package ssh

import (
	"crypto/dsa"

	gossh "golang.org/x/crypto/ssh"
)

type dsaKeyPair struct {
	privateKey      *dsa.PrivateKey
	publicKey       gossh.PublicKey
	name            string
	privatePemBlock []byte
}

func (o dsaKeyPair) Type() KeyPairType {
	return Dsa
}

func (o dsaKeyPair) Bits() int {
	return 1024
}

func (o dsaKeyPair) Name() string {
	return o.name
}

func (o dsaKeyPair) Description() string {
	return description(o)
}

func (o dsaKeyPair) PrivateKeyPemBlock() []byte {
	return o.privatePemBlock
}

func (o dsaKeyPair) PublicKeyAuthorizedKeysLine(nl NewLineOption) []byte {
	return publicKeyAuthorizedKeysLine(o.publicKey, o.name, nl)
}
