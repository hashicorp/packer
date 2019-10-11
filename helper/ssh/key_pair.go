package ssh

import (
	"bytes"
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	gossh "golang.org/x/crypto/ssh"
)

const (
	// defaultRsaBits is the default bits of entropy for a new RSA
	// key pair. That's a lot of bits.
	defaultRsaBits = 4096

	// Markers for various SSH key pair types.
	Default KeyPairType = ""
	Rsa     KeyPairType = "RSA"
	Ecdsa   KeyPairType = "ECDSA"
	Dsa     KeyPairType = "DSA"
	Ed25519 KeyPairType = "ED25519"
)

// KeyPairType represents different types of SSH key pairs.
// See the 'const' block for details.
type KeyPairType string

func (o KeyPairType) String() string {
	return string(o)
}

// KeyPair represents an SSH key pair.
type KeyPair struct {
	// PrivateKeyPemBlock represents the key pair's private key in
	// ASN.1 Distinguished Encoding Rules (DER) format in a
	// Privacy-Enhanced Mail (PEM) block.
	PrivateKeyPemBlock []byte

	// PublicKeyAuthorizedKeysLine represents the key pair's public key
	// as a line in OpenSSH authorized_keys.
	PublicKeyAuthorizedKeysLine []byte

	// Comment is the key pair's comment. This is typically used
	// to identify the key pair's owner in the SSH user's
	// 'authorized_keys' file.
	Comment string
}

// KeyPairFromPrivateKey returns a KeyPair loaded from an existing private key.
//
// Supported key pair types include:
// 	- DSA
// 	- ECDSA
// 	- ED25519
// 	- RSA
func KeyPairFromPrivateKey(config FromPrivateKeyConfig) (KeyPair, error) {
	privateKey, err := gossh.ParseRawPrivateKey(config.RawPrivateKeyPemBlock)
	if err != nil {
		return KeyPair{}, err
	}

	switch pk := privateKey.(type) {
	case crypto.Signer:
		// crypto.Signer is implemented by ecdsa.PrivateKey,
		// ed25519.PrivateKey, and rsa.PrivateKey - separate cases
		// for each PrivateKey type would be redundant.
		publicKey, err := gossh.NewPublicKey(pk.Public())
		if err != nil {
			return KeyPair{}, err
		}
		return KeyPair{
			Comment:                     config.Comment,
			PrivateKeyPemBlock:          config.RawPrivateKeyPemBlock,
			PublicKeyAuthorizedKeysLine: authorizedKeysLine(publicKey, config.Comment),
		}, nil
	case *dsa.PrivateKey:
		publicKey, err := gossh.NewPublicKey(&pk.PublicKey)
		if err != nil {
			return KeyPair{}, err
		}
		return KeyPair{
			Comment:                     config.Comment,
			PrivateKeyPemBlock:          config.RawPrivateKeyPemBlock,
			PublicKeyAuthorizedKeysLine: authorizedKeysLine(publicKey, config.Comment),
		}, nil
	}

	return KeyPair{}, fmt.Errorf("Cannot parse existing SSH key pair - unknown key pair type")
}

// FromPrivateKeyConfig describes how an SSH key pair should be loaded from an
// existing private key.
type FromPrivateKeyConfig struct {
	// RawPrivateKeyPemBlock is the raw private key that the key pair
	// should be loaded from.
	RawPrivateKeyPemBlock []byte

	// Comment is the key pair's comment. This is typically used
	// to identify the key pair's owner in the SSH user's
	// 'authorized_keys' file.
	Comment string
}

// NewKeyPair generates a new SSH key pair using the specified
// CreateKeyPairConfig.
func NewKeyPair(config CreateKeyPairConfig) (KeyPair, error) {
	if config.Type == Default {
		config.Type = Ecdsa
	}

	switch config.Type {
	case Ecdsa:
		return newEcdsaKeyPair(config)
	case Rsa:
		return newRsaKeyPair(config)
	}

	return KeyPair{}, fmt.Errorf("Unable to generate new key pair, type %s is not supported",
		config.Type.String())
}

// newEcdsaKeyPair returns a new ECDSA SSH key pair.
func newEcdsaKeyPair(config CreateKeyPairConfig) (KeyPair, error) {
	var curve elliptic.Curve

	switch config.Bits {
	case 0:
		config.Bits = 521
		fallthrough
	case 521:
		curve = elliptic.P521()
	case 384:
		curve = elliptic.P384()
	case 256:
		curve = elliptic.P256()
	case 224:
		// Not supported by "golang.org/x/crypto/ssh".
		return KeyPair{}, fmt.Errorf("golang.org/x/crypto/ssh does not support %d bits", config.Bits)
	default:
		return KeyPair{}, fmt.Errorf("crypto/elliptic does not support %d bits", config.Bits)
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return KeyPair{}, err
	}

	sshPublicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return KeyPair{}, err
	}

	privateRaw, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return KeyPair{}, err
	}

	privatePem, err := rawPemBlock(&pem.Block{
		Type:    "EC PRIVATE KEY",
		Headers: nil,
		Bytes:   privateRaw,
	})
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		PrivateKeyPemBlock:          privatePem,
		PublicKeyAuthorizedKeysLine: authorizedKeysLine(sshPublicKey, config.Comment),
		Comment:                     config.Comment,
	}, nil
}

// newRsaKeyPair returns a new RSA SSH key pair.
func newRsaKeyPair(config CreateKeyPairConfig) (KeyPair, error) {
	if config.Bits == 0 {
		config.Bits = defaultRsaBits
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, config.Bits)
	if err != nil {
		return KeyPair{}, err
	}

	sshPublicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return KeyPair{}, err
	}

	privatePemBlock, err := rawPemBlock(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		PrivateKeyPemBlock:          privatePemBlock,
		PublicKeyAuthorizedKeysLine: authorizedKeysLine(sshPublicKey, config.Comment),
		Comment:                     config.Comment,
	}, nil
}

// CreateKeyPairConfig describes how an SSH key pair should be created.
type CreateKeyPairConfig struct {
	// Type describes the key pair's type.
	Type KeyPairType

	// Bits represents the key pair's bits of entropy. E.g., 4096 for
	// a 4096 bit RSA key pair, or 521 for a ECDSA key pair with a
	// 521-bit curve.
	Bits int

	// Comment is the resulting key pair's comment. This is typically
	// used to identify the key pair's owner in the SSH user's
	// 'authorized_keys' file.
	Comment string
}

// rawPemBlock encodes a pem.Block to a slice of bytes.
func rawPemBlock(block *pem.Block) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)

	err := pem.Encode(buffer, block)
	if err != nil {
		return []byte{}, err
	}

	return buffer.Bytes(), nil
}

// authorizedKeysLine serializes key for inclusion in an OpenSSH
// authorized_keys file. The return value ends without newline so
// a comment can be appended to the end.
func authorizedKeysLine(key gossh.PublicKey, comment string) []byte {
	marshaledPublicKey := gossh.MarshalAuthorizedKey(key)

	// Remove the mandatory unix new line. Awful, but the go
	// ssh library automatically appends a unix new line.
	// We remove it so a key comment can be safely appended to the
	// end of the string.
	marshaledPublicKey = bytes.TrimSpace(marshaledPublicKey)

	if len(strings.TrimSpace(comment)) > 0 {
		marshaledPublicKey = append(marshaledPublicKey, ' ')
		marshaledPublicKey = append(marshaledPublicKey, comment...)
	}

	return marshaledPublicKey
}
