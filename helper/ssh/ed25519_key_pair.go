package ssh

import (
	"golang.org/x/crypto/ed25519"

	gossh "golang.org/x/crypto/ssh"
)

type ed25519KeyPair struct {
	privateKey      *ed25519.PrivateKey
	publicKey       gossh.PublicKey
	name            string
	privatePemBlock []byte
}

func (o ed25519KeyPair) Type() KeyPairType {
	return Ed25519
}

func (o ed25519KeyPair) Bits() int {
	return 256
}

func (o ed25519KeyPair) Name() string {
	return o.name
}

func (o ed25519KeyPair) Description() string {
	return description(o)
}

func (o ed25519KeyPair) PrivateKeyPemBlock() []byte {
	return o.privatePemBlock
}

func (o ed25519KeyPair) PublicKeyAuthorizedKeysLine(nl NewLineOption) []byte {
	return publicKeyAuthorizedKeysLine(o.publicKey, o.name, nl)
}
