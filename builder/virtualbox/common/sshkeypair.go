package common

// TODO: Make this available to other packer APIs.
//  Perhaps through 'helper/ssh'?

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

	"golang.org/x/crypto/ssh"
)

const (
	// That's a lot of bits.
	defaultRsaBits = 4096

	// rsaSsh is a SSH key pair of RSA type.
	rsaSsh sshKeyPairType = "rsa"

	// ecdsaSsh is a SSH key pair of ECDSA type.
	ecdsaSsh sshKeyPairType = "ecdsa"
)

// sshKeyPairType represents different types of SSH key pairs.
// For example, RSA.
type sshKeyPairType string

func (o sshKeyPairType) String() string {
	return string(o)
}

const (
	// unixNewLine is a unix new line.
	unixNewLine newLineOption = "\n"

	// windowsNewLine is a Windows new line.
	windowsNewLine newLineOption = "\r\n"

	// noNewLine will not append a new line.
	noNewLine newLineOption = ""
)

// newLineOption specifies the type of new line to append to a string.
// See the 'const' block for choices.
type newLineOption string

func (o newLineOption) String() string {
	return string(o)
}

func (o newLineOption) Bytes() []byte {
	return []byte(o)
}

// sshKeyPairBuilder builds SSH key pairs.
type sshKeyPairBuilder interface {
	// SetType sets the key pair type.
	SetType(sshKeyPairType) sshKeyPairBuilder

	// SetBits sets the key pair's bits of entropy.
	SetBits(int) sshKeyPairBuilder

	// SetName sets the name of the key pair. This is primarily used
	// to identify the public key in the authorized_keys file.
	SetName(string) sshKeyPairBuilder

	// Build returns a SSH key pair.
	//
	// The following defaults are used if not specified:
	//	Default type: ECDSA
	//	Default bits of entropy:
	//		- RSA: 4096
	//		- ECDSA: 521
	// 	Default name: (empty string)
	Build() (sshKeyPair, error)
}

type defaultSshKeyPairBuilder struct {
	// kind describes the resulting key pair's type.
	kind sshKeyPairType

	// bits is the resulting key pair's bits of entropy.
	bits int

	// name is the resulting key pair's name.
	name string
}

func (o *defaultSshKeyPairBuilder) SetType(kind sshKeyPairType) sshKeyPairBuilder {
	o.kind = kind
	return o
}

func (o *defaultSshKeyPairBuilder) SetBits(bits int) sshKeyPairBuilder {
	o.bits = bits
	return o
}

func (o *defaultSshKeyPairBuilder) SetName(name string) sshKeyPairBuilder {
	o.name = name
	return o
}

func (o *defaultSshKeyPairBuilder) Build() (sshKeyPair, error) {
	switch o.kind {
	case rsaSsh:
		return o.newRsaSshKeyPair()
	case ecdsaSsh:
		// Default case.
	}

	return o.newEcdsaSshKeyPair()
}

// sshKeyPair represents a SSH key pair.
type sshKeyPair interface {
	// Type returns the key pair's type.
	Type() sshKeyPairType

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
	PublicKeyAuthorizedKeysFormat(newLineOption) []byte
}

type defaultSshKeyPair struct {
	// kind is the key pair's type.
	kind sshKeyPairType

	// bits is the key pair's bits of entropy.
	bits int

	// name is the key pair's name.
	name string

	// privateKeyDerBytes is the private key's bytes
	// in ASN.1 DER format
	privateKeyDerBytes []byte

	// publicKey is the key pair's public key.
	publicKey ssh.PublicKey
}

func (o defaultSshKeyPair) Type() sshKeyPairType {
	return o.kind
}

func (o defaultSshKeyPair) Bits() int {
	return o.bits
}

func (o defaultSshKeyPair) Name() string {
	return o.name
}

func (o defaultSshKeyPair) Description() string {
	return o.kind.String() + " " + strconv.Itoa(o.bits)
}

func (o defaultSshKeyPair) PrivateKeyPemBlock() []byte {
	t := "UNKNOWN PRIVATE KEY"

	switch o.kind {
	case ecdsaSsh:
		t = "EC PRIVATE KEY"
	case rsaSsh:
		t = "RSA PRIVATE KEY"
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:    t,
		Headers: nil,
		Bytes:   o.privateKeyDerBytes,
	})
}

func (o defaultSshKeyPair) PublicKeyAuthorizedKeysFormat(nl newLineOption) []byte {
	result := ssh.MarshalAuthorizedKey(o.publicKey)

	if len(strings.TrimSpace(o.name)) > 0 {
		// Awful, but the go ssh library automatically appends
		// a unix new line.
		result = bytes.TrimSuffix(result, unixNewLine.Bytes())
		result = append(result, ' ')
		result = append(result, o.name...)
	}

	switch nl {
	case noNewLine:
		result = bytes.TrimSuffix(result, unixNewLine.Bytes())
	case windowsNewLine:
		result = bytes.TrimSuffix(result, unixNewLine.Bytes())
		result = append(result, nl.Bytes()...)
	case unixNewLine:
		fallthrough
	default:
		// This is how all the other "SSH key pair" code works in
		// the different builders.
		if !bytes.HasSuffix(result, unixNewLine.Bytes()) {
			result = append(result, unixNewLine.Bytes()...)
		}
	}

	return result
}

// newEcdsaSshKeyPair returns a new ECDSA SSH key pair for the given bits
// of entropy.
func (o *defaultSshKeyPairBuilder) newEcdsaSshKeyPair() (sshKeyPair, error) {
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
		return &defaultSshKeyPair{}, errors.New("golang.org/x/crypto/ssh does not support " +
				strconv.Itoa(o.bits) + " bits")
	default:
		return &defaultSshKeyPair{}, errors.New("crypto/elliptic does not support " +
			strconv.Itoa(o.bits) + " bits")
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return &defaultSshKeyPair{}, err
	}

	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return &defaultSshKeyPair{}, err
	}

	raw, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return &defaultSshKeyPair{}, err
	}

	return &defaultSshKeyPair{
		kind:               ecdsaSsh,
		bits:               o.bits,
		name:               o.name,
		privateKeyDerBytes: raw,
		publicKey:          sshPublicKey,
	}, nil
}

// newRsaSshKeyPair returns a new RSA SSH key pair for the given bits
// of entropy.
func (o *defaultSshKeyPairBuilder) newRsaSshKeyPair() (sshKeyPair, error) {
	if o.bits == 0 {
		o.bits = defaultRsaBits
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, o.bits)
	if err != nil {
		return &defaultSshKeyPair{}, err
	}

	sshPublicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return &defaultSshKeyPair{}, err
	}

	return &defaultSshKeyPair{
		kind:               rsaSsh,
		bits:               o.bits,
		name:               o.name,
		privateKeyDerBytes: x509.MarshalPKCS1PrivateKey(privateKey),
		publicKey:          sshPublicKey,
	}, nil
}

func newSshKeyPairBuilder() sshKeyPairBuilder {
	return &defaultSshKeyPairBuilder{}
}
