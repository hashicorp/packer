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
	Rsa KeyPairType = "rsa"

	// Ecdsa is a SSH key pair of ECDSA type.
	Ecdsa KeyPairType = "ecdsa"
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
		return o.newRsaSshKeyPair()
	case Ecdsa:
		// Default case.
	}

	return o.newEcdsaSshKeyPair()
}

// newEcdsaSshKeyPair returns a new ECDSA SSH key pair.
func (o *defaultKeyPairBuilder) newEcdsaSshKeyPair() (KeyPair, error) {
	var curve elliptic.Curve

	switch o.bits {
	case 0:
		o.bits = 521
		fallthrough
	case 521:
		curve = elliptic.P521()
	case 384:
		elliptic.P384()
	case 256:
		elliptic.P256()
	case 224:
		// Not supported by "golang.org/x/crypto/ssh".
		return &defaultKeyPair{}, errors.New("golang.org/x/crypto/ssh does not support " +
			strconv.Itoa(o.bits) + " bits")
	default:
		return &defaultKeyPair{}, errors.New("crypto/elliptic does not support " +
			strconv.Itoa(o.bits) + " bits")
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return &defaultKeyPair{}, err
	}

	sshPublicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return &defaultKeyPair{}, err
	}

	raw, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return &defaultKeyPair{}, err
	}

	return &defaultKeyPair{
		kind:               Ecdsa,
		bits:               o.bits,
		name:               o.name,
		privateKeyDerBytes: raw,
		publicKey:          sshPublicKey,
	}, nil
}

// newRsaSshKeyPair returns a new RSA SSH key pair.
func (o *defaultKeyPairBuilder) newRsaSshKeyPair() (KeyPair, error) {
	if o.bits == 0 {
		o.bits = defaultRsaBits
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, o.bits)
	if err != nil {
		return &defaultKeyPair{}, err
	}

	sshPublicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return &defaultKeyPair{}, err
	}

	return &defaultKeyPair{
		kind:               Rsa,
		bits:               o.bits,
		name:               o.name,
		privateKeyDerBytes: x509.MarshalPKCS1PrivateKey(privateKey),
		publicKey:          sshPublicKey,
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

	// PublicKeyAuthorizedKeysFormat returns a slice of bytes
	// representing the public key in OpenSSH authorized_keys format
	// with the specified new line.
	PublicKeyAuthorizedKeysFormat(NewLineOption) []byte
}

type defaultKeyPair struct {
	// kind is the key pair's type.
	kind KeyPairType

	// bits is the key pair's bits of entropy.
	bits int

	// name is the key pair's name.
	name string

	// privateKeyDerBytes is the private key's bytes
	// in ASN.1 DER format
	privateKeyDerBytes []byte

	// publicKey is the key pair's public key.
	publicKey gossh.PublicKey
}

func (o defaultKeyPair) Type() KeyPairType {
	return o.kind
}

func (o defaultKeyPair) Bits() int {
	return o.bits
}

func (o defaultKeyPair) Name() string {
	return o.name
}

func (o defaultKeyPair) Description() string {
	return o.kind.String() + " " + strconv.Itoa(o.bits)
}

func (o defaultKeyPair) PrivateKeyPemBlock() []byte {
	t := "UNKNOWN PRIVATE KEY"

	switch o.kind {
	case Ecdsa:
		t = "EC PRIVATE KEY"
	case Rsa:
		t = "RSA PRIVATE KEY"
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:    t,
		Headers: nil,
		Bytes:   o.privateKeyDerBytes,
	})
}

func (o defaultKeyPair) PublicKeyAuthorizedKeysFormat(nl NewLineOption) []byte {
	result := gossh.MarshalAuthorizedKey(o.publicKey)

	if len(strings.TrimSpace(o.name)) > 0 {
		// Awful, but the go ssh library automatically appends
		// a unix new line.
		result = bytes.TrimSuffix(result, UnixNewLine.Bytes())
		result = append(result, ' ')
		result = append(result, o.name...)
	}

	switch nl {
	case NoNewLine:
		result = bytes.TrimSuffix(result, UnixNewLine.Bytes())
	case WindowsNewLine:
		result = bytes.TrimSuffix(result, UnixNewLine.Bytes())
		result = append(result, nl.Bytes()...)
	case UnixNewLine:
		fallthrough
	default:
		// This is how all the other "SSH key pair" code works in
		// the different builders.
		if !bytes.HasSuffix(result, UnixNewLine.Bytes()) {
			result = append(result, UnixNewLine.Bytes()...)
		}
	}

	return result
}

func NewKeyPairBuilder() KeyPairBuilder {
	return &defaultKeyPairBuilder{}
}
