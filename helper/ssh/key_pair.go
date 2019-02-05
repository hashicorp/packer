package ssh

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strconv"
	"strings"

	gossh "golang.org/x/crypto/ssh"
)

const (
	// That's a lot of bits.
	defaultRsaBits = 4096

	// Rsa is a SSH key pair of RSA type.
	Rsa KeyPairType = "RSA"

	// Ecdsa is a SSH key pair of ECDSA type.
	Ecdsa KeyPairType = "ECDSA"
)

// KeyPairType represents different types of SSH key pairs.
// See the 'const' block for details.
type KeyPairType string

func (o KeyPairType) String() string {
	return string(o)
}

const (
	// UnixNewLine is a unix new line.
	UnixNewLine NewLineOption = "\n"

	// WindowsNewLine is a Windows new line.
	WindowsNewLine NewLineOption = "\r\n"

	// NoNewLine will not append a new line.
	NoNewLine NewLineOption = ""
)

// NewLineOption specifies the type of new line to append to a string.
// See the 'const' block for choices.
type NewLineOption string

func (o NewLineOption) String() string {
	return string(o)
}

func (o NewLineOption) Bytes() []byte {
	return []byte(o)
}

// KeyPairBuilder builds SSH key pairs.
type KeyPairBuilder interface {
	// SetType sets the key pair type.
	SetType(KeyPairType) KeyPairBuilder

	// SetBits sets the key pair's bits of entropy.
	SetBits(int) KeyPairBuilder

	// SetName sets the name of the key pair. This is primarily used
	// to identify the public key in the authorized_keys file.
	SetName(string) KeyPairBuilder

	// Build returns a SSH key pair.
	//
	// The following defaults are used if not specified:
	//	Default type: ECDSA
	//	Default bits of entropy:
	//		- RSA: 4096
	//		- ECDSA: 521
	// 	Default name: (empty string)
	Build() (KeyPair, error)
}

type defaultKeyPairBuilder struct {
	// kind describes the resulting key pair's type.
	kind KeyPairType

	// bits is the resulting key pair's bits of entropy.
	bits int

	// name is the resulting key pair's name.
	name string
}

func (o *defaultKeyPairBuilder) SetType(kind KeyPairType) KeyPairBuilder {
	o.kind = kind
	return o
}

func (o *defaultKeyPairBuilder) SetBits(bits int) KeyPairBuilder {
	o.bits = bits
	return o
}

func (o *defaultKeyPairBuilder) SetName(name string) KeyPairBuilder {
	o.name = name
	return o
}

func (o *defaultKeyPairBuilder) Build() (KeyPair, error) {
	switch o.kind {
	case Rsa:
		return o.newRsaKeyPair()
	case Ecdsa:
		// Default case.
	}

	return o.newEcdsaKeyPair()
}

// newEcdsaKeyPair returns a new ECDSA SSH key pair.
func (o *defaultKeyPairBuilder) newEcdsaKeyPair() (KeyPair, error) {
	var curve elliptic.Curve

	switch o.bits {
	case 0:
		o.bits = 521
		fallthrough
	case 521:
		curve = elliptic.P521()
	case 384:
		curve = elliptic.P384()
	case 256:
		curve = elliptic.P256()
	case 224:
		// Not supported by "golang.org/x/crypto/ssh".
		return &ecdsaKeyPair{}, errors.New("golang.org/x/crypto/ssh does not support " +
			strconv.Itoa(o.bits) + " bits")
	default:
		return &ecdsaKeyPair{}, errors.New("crypto/elliptic does not support " +
			strconv.Itoa(o.bits) + " bits")
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return &ecdsaKeyPair{}, err
	}

	sshPublicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return &ecdsaKeyPair{}, err
	}

	privateRaw, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return &ecdsaKeyPair{}, err
	}

	privatePem, err := rawPemBlock(&pem.Block{
		Type:    "EC PRIVATE KEY",
		Headers: nil,
		Bytes:   privateRaw,
	})
	if err != nil {
		return &ecdsaKeyPair{}, err
	}

	return &ecdsaKeyPair{
		privateKey:      privateKey,
		publicKey:       sshPublicKey,
		name:            o.name,
		privatePemBlock: privatePem,
	}, nil
}

// newRsaKeyPair returns a new RSA SSH key pair.
func (o *defaultKeyPairBuilder) newRsaKeyPair() (KeyPair, error) {
	if o.bits == 0 {
		o.bits = defaultRsaBits
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, o.bits)
	if err != nil {
		return &rsaKeyPair{}, err
	}

	sshPublicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return &rsaKeyPair{}, err
	}

	privatePemBlock, err := rawPemBlock(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return &rsaKeyPair{}, err
	}

	return &rsaKeyPair{
		privateKey:      privateKey,
		publicKey:       sshPublicKey,
		name:            o.name,
		privatePemBlock: privatePemBlock,
	}, nil
}

// KeyPair represents a SSH key pair.
type KeyPair interface {
	// Type returns the key pair's type.
	Type() KeyPairType

	// Bits returns the bits of entropy.
	Bits() int

	// Name returns the key pair's name. An empty string is
	// returned is no name was specified.
	Name() string

	// Description returns a brief description of the key pair that
	// is suitable for log messages or printing.
	Description() string

	// PrivateKeyPemBlock returns a slice of bytes representing
	// the private key in ASN.1 Distinguished Encoding Rules (DER)
	// format in a Privacy-Enhanced Mail (PEM) block.
	PrivateKeyPemBlock() []byte

	// PublicKeyAuthorizedKeysLine returns a slice of bytes
	// representing the public key as a line in OpenSSH authorized_keys
	// format with the specified new line.
	PublicKeyAuthorizedKeysLine(NewLineOption) []byte
}

func NewKeyPairBuilder() KeyPairBuilder {
	return &defaultKeyPairBuilder{}
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

// description returns a string describing a key pair.
func description(kp KeyPair) string {
	buffer := bytes.NewBuffer(nil)

	buffer.WriteString(strconv.Itoa(kp.Bits()))
	buffer.WriteString(" bit ")
	buffer.WriteString(kp.Type().String())

	if len(kp.Name()) > 0 {
		buffer.WriteString(" named ")
		buffer.WriteString(kp.Name())
	}

	return buffer.String()
}

// publicKeyAuthorizedKeysLine returns a slice of bytes representing a SSH
// public key as a line in OpenSSH authorized_keys format.
func publicKeyAuthorizedKeysLine(publicKey gossh.PublicKey, name string, nl NewLineOption) []byte {
	result := gossh.MarshalAuthorizedKey(publicKey)

	// Remove the mandatory unix new line.
	// Awful, but the go ssh library automatically appends
	// a unix new line.
	result = bytes.TrimSpace(result)

	if len(strings.TrimSpace(name)) > 0 {
		result = append(result, ' ')
		result = append(result, name...)
	}

	switch nl {
	case WindowsNewLine:
		result = append(result, nl.Bytes()...)
	case UnixNewLine:
		// This is how all the other "SSH key pair" code works in
		// the different builders.
		result = append(result, UnixNewLine.Bytes()...)
	}

	return result
}
