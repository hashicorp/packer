package ssh

import (
	"crypto/rsa"

	gossh "golang.org/x/crypto/ssh"
)

type rsaKeyPair struct {
	privateKey      *rsa.PrivateKey
	publicKey       gossh.PublicKey
	name            string
	privatePemBlock []byte
}

func (o rsaKeyPair) Type() KeyPairType {
	return Rsa
}

func (o rsaKeyPair) Bits() int {
	return o.privateKey.N.BitLen()
}

func (o rsaKeyPair) Name() string {
	return o.name
}

func (o rsaKeyPair) Description() string {
	return description(o)
}

func (o rsaKeyPair) PrivateKeyPemBlock() []byte {
	return o.privatePemBlock
}

func (o rsaKeyPair) PublicKeyAuthorizedKeysLine(nl NewLineOption) []byte {
	return publicKeyAuthorizedKeysLine(o.publicKey, o.name, nl)
}
