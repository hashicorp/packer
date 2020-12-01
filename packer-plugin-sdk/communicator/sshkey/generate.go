package sshkey

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/ssh"
)

type Algorithm int

//go:generate enumer -type Algorithm -transform snake
const (
	RSA Algorithm = iota
	DSA
	ECDSA
	ED25519
)

var (
	ErrUnknownAlgorithm    = fmt.Errorf("sshkey: unknown private key algorithm")
	ErrInvalidRSAKeySize   = fmt.Errorf("sshkey: invalid private key rsa size: must be more than 1024")
	ErrInvalidECDSAKeySize = fmt.Errorf("sshkey: invalid private key ecdsa size, must be one of 256, 384 or 521")
	ErrInvalidDSAKeySize   = fmt.Errorf("sshkey: invalid private key dsa size, must be one of 1024, 2048 or 3072")
)

// Pair represents an ssh key pair, as in
type Pair struct {
	Private []byte
	Public  []byte
}

func NewPair(public, private interface{}) (*Pair, error) {
	kb, err := x509.MarshalPKCS8PrivateKey(private)
	if err != nil {
		return nil, err
	}

	privBlk := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   kb,
	}

	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, err
	}
	return &Pair{
		Private: pem.EncodeToMemory(privBlk),
		Public:  ssh.MarshalAuthorizedKey(publicKey),
	}, nil
}

// PairFromED25519 marshalls a valid pair of openssh pem for ED25519 keypairs.
// NewPair can handle ed25519 pairs but generates the wrong format apparently:
// `Load key "id_ed25519": invalid format` is the error that happens when I try
// to ssh with such a key.
func PairFromED25519(public ed25519.PublicKey, private ed25519.PrivateKey) (*Pair, error) {
	// see https://github.com/golang/crypto/blob/7f63de1d35b0f77fa2b9faea3e7deb402a2383c8/ssh/keys.go#L1273-L1443
	key := struct {
		Pub     []byte
		Priv    []byte
		Comment string
		Pad     []byte `ssh:"rest"`
	}{
		Pub:  public,
		Priv: private,
	}
	keyBytes := ssh.Marshal(key)

	pk1 := struct {
		Check1  uint32
		Check2  uint32
		Keytype string
		Rest    []byte `ssh:"rest"`
	}{
		Keytype: ssh.KeyAlgoED25519,
		Rest:    keyBytes,
	}
	pk1Bytes := ssh.Marshal(pk1)

	k := struct {
		CipherName   string
		KdfName      string
		KdfOpts      string
		NumKeys      uint32
		PubKey       []byte
		PrivKeyBlock []byte
	}{
		CipherName:   "none",
		KdfName:      "none",
		KdfOpts:      "",
		NumKeys:      1,
		PrivKeyBlock: pk1Bytes,
	}

	const opensshV1Magic = "openssh-key-v1\x00"

	privBlk := &pem.Block{
		Type:    "OPENSSH PRIVATE KEY",
		Headers: nil,
		Bytes:   append([]byte(opensshV1Magic), ssh.Marshal(k)...),
	}
	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, err
	}
	return &Pair{
		Private: pem.EncodeToMemory(privBlk),
		Public:  ssh.MarshalAuthorizedKey(publicKey),
	}, nil
}

// PairFromDSA marshalls a valid pair of openssh pem for dsa keypairs.
// x509.MarshalPKCS8PrivateKey does not know how to deal with dsa keys.
func PairFromDSA(key *dsa.PrivateKey) (*Pair, error) {
	// see https://github.com/golang/crypto/blob/7f63de1d35b0f77fa2b9faea3e7deb402a2383c8/ssh/keys.go#L1186-L1195
	// and https://linux.die.net/man/1/dsa
	k := struct {
		Version int
		P       *big.Int
		Q       *big.Int
		G       *big.Int
		Pub     *big.Int
		Priv    *big.Int
	}{
		Version: 0,
		P:       key.P,
		Q:       key.Q,
		G:       key.G,
		Pub:     key.Y,
		Priv:    key.X,
	}
	kb, err := asn1.Marshal(k)
	if err != nil {
		return nil, err
	}
	privBlk := &pem.Block{
		Type:    "DSA PRIVATE KEY",
		Headers: nil,
		Bytes:   kb,
	}
	publicKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	return &Pair{
		Private: pem.EncodeToMemory(privBlk),
		Public:  ssh.MarshalAuthorizedKey(publicKey),
	}, nil
}

// GeneratePair generates a Private/Public key pair using algorithm t.
//
// When rand is nil "crypto/rand".Reader will be used.
//
// bits specifies the number of bits in the key to create. For RSA keys, the
// minimum size is 1024 bits and the default is 3072 bits. Generally, 3072 bits
// is considered sufficient. DSA keys must be exactly 1024 bits - or 2 or 3
// times that - as specified by FIPS 186-2. For ECDSA keys, bits determines the
// key length by selecting from one of three elliptic curve sizes: 256, 384 or
// 521 bits. Attempting to use bit lengths other than these three values for
// ECDSA keys will fail. Ed25519 keys have a fixed length and the bits will
// be ignored.
func GeneratePair(t Algorithm, rand io.Reader, bits int) (*Pair, error) {
	if rand == nil {
		rand = cryptorand.Reader
	}
	switch t {
	case DSA:
		if bits == 0 {
			// currently the ssh package can only decode 1024 bits dsa keys, so
			// that's going be the default for now see
			// https://github.com/golang/crypto/blob/7f63de1d35b0f77fa2b9faea3e7deb402a2383c8/ssh/keys.go#L411-L420
			bits = 1024
		}
		var sizes dsa.ParameterSizes
		switch bits {
		case 1024:
			sizes = dsa.L1024N160
		case 2048:
			sizes = dsa.L2048N256
		case 3072:
			sizes = dsa.L3072N256
		default:
			return nil, ErrInvalidDSAKeySize
		}

		params := dsa.Parameters{}
		if err := dsa.GenerateParameters(&params, rand, sizes); err != nil {
			return nil, err
		}

		dsakey := &dsa.PrivateKey{
			PublicKey: dsa.PublicKey{
				Parameters: params,
			},
		}
		if err := dsa.GenerateKey(dsakey, rand); err != nil {
			return nil, err
		}
		return PairFromDSA(dsakey)
	case ECDSA:
		if bits == 0 {
			bits = 521
		}
		var ecdsakey *ecdsa.PrivateKey
		var err error
		switch bits {
		case 256:
			ecdsakey, err = ecdsa.GenerateKey(elliptic.P256(), rand)
		case 384:
			ecdsakey, err = ecdsa.GenerateKey(elliptic.P384(), rand)
		case 521:
			ecdsakey, err = ecdsa.GenerateKey(elliptic.P521(), rand)
		default:
			ecdsakey, err = nil, ErrInvalidECDSAKeySize
		}
		if err != nil {
			return nil, err
		}
		return NewPair(&ecdsakey.PublicKey, ecdsakey)
	case ED25519:
		publicKey, privateKey, err := ed25519.GenerateKey(rand)
		if err != nil {
			return nil, err
		}
		return PairFromED25519(publicKey, privateKey)
	case RSA:
		if bits == 0 {
			bits = 4096
		}
		if bits < 1024 {
			return nil, ErrInvalidRSAKeySize
		}
		rsakey, err := rsa.GenerateKey(rand, bits)
		if err != nil {
			return nil, err
		}
		return NewPair(&rsakey.PublicKey, rsakey)
	default:
		return nil, ErrUnknownAlgorithm
	}
}
