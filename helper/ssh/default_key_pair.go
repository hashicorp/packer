package ssh

type defaultKeyPair struct {
}

func (o defaultKeyPair) Type() KeyPairType {
	return  Default
}

func (o defaultKeyPair) Bits() int {
	return 0
}

func (o defaultKeyPair) Name() string {
	return ""
}

func (o defaultKeyPair) Description() string {
	return ""
}

func (o defaultKeyPair) PrivateKeyPemBlock() []byte {
	return []byte{}
}

func (o defaultKeyPair) PublicKeyAuthorizedKeysLine(nl NewLineOption) []byte {
	return []byte{}
}
