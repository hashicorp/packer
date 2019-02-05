package ssh

import (
	"crypto/ecdsa"

	gossh "golang.org/x/crypto/ssh"
)

type ecdsaKeyPair struct {
	privateKey      *ecdsa.PrivateKey
	publicKey       gossh.PublicKey
	name            string
	privatePemBlock []byte
}

func (o ecdsaKeyPair) Type() KeyPairType {
	return Ecdsa
}

func (o ecdsaKeyPair) Bits() int {
	return o.privateKey.Curve.Params().BitSize
}

func (o ecdsaKeyPair) Name() string {
	return o.name
}

func (o ecdsaKeyPair) Description() string {
	return description(o)
}

func (o ecdsaKeyPair) PrivateKeyPemBlock() []byte {
	return o.privatePemBlock
}

func (o ecdsaKeyPair) PublicKeyAuthorizedKeysLine(nl NewLineOption) []byte {
	return publicKeyAuthorizedKeysLine(o.publicKey, o.name, nl)
}
