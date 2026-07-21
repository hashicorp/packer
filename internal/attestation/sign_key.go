// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
)

func init() {
	RegisterSigner(SigningModeKey, newPEMSigner)
}

type pemSigner struct {
	signer   crypto.Signer
	verifier *pemVerifier
}

type pemVerifier struct {
	publicKey crypto.PublicKey
	keyID     string
}

func newPEMSigner(_ context.Context, cfg BackendConfig) (Signer, error) {
	if cfg.SignerRef == "" {
		return nil, fmt.Errorf("signing_mode %q requires signer", SigningModeKey)
	}

	signer, verifier, err := loadPEMSigner(cfg.SignerRef)
	if err != nil {
		return nil, err
	}

	return &pemSigner{signer: signer, verifier: verifier}, nil
}

func (s *pemSigner) Sign(_ context.Context, payloadType string, payload []byte) (Signature, error) {
	pae := PreAuthEncode(payloadType, payload)

	var message []byte
	var opts crypto.SignerOpts
	if _, ok := s.signer.Public().(ed25519.PublicKey); ok {
		message = pae
		opts = crypto.Hash(0)
	} else {
		digest := sha256.Sum256(pae)
		message = digest[:]
		opts = crypto.SHA256
	}

	signature, err := s.signer.Sign(rand.Reader, message, opts)
	if err != nil {
		return Signature{}, fmt.Errorf("sign payload: %w", err)
	}

	return Signature{
		KeyID: s.verifier.KeyID(),
		Sig:   signature,
	}, nil
}

func (s *pemSigner) Verifier(context.Context, BackendConfig) (Verifier, error) {
	return s.verifier, nil
}

func (v *pemVerifier) Verify(_ context.Context, payloadType string, payload, signature []byte) error {
	pae := PreAuthEncode(payloadType, payload)

	switch publicKey := v.publicKey.(type) {
	case *rsa.PublicKey:
		digest := sha256.Sum256(pae)
		return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature)
	case *ecdsa.PublicKey:
		digest := sha256.Sum256(pae)
		if !ecdsa.VerifyASN1(publicKey, digest[:], signature) {
			return fmt.Errorf("ECDSA verification failed")
		}
		return nil
	case ed25519.PublicKey:
		if !ed25519.Verify(publicKey, pae, signature) {
			return fmt.Errorf("Ed25519 verification failed")
		}
		return nil
	default:
		return fmt.Errorf("unsupported public key type %T", v.publicKey)
	}
}

func (v *pemVerifier) KeyID() string {
	return v.keyID
}

func LoadPEMVerifier(path string) (Verifier, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read verifier %q: %w", path, err)
	}

	publicKey, rawVerifier, err := loadPEMPublicKey(contents)
	if err != nil {
		return nil, fmt.Errorf("load verifier %q: %w", path, err)
	}

	return &pemVerifier{
		publicKey: publicKey,
		keyID:     sha256Hex(rawVerifier),
	}, nil
}

func LoadPEMVerifierBytes(contents []byte) (*pemVerifier, error) {
	publicKey, rawVerifier, err := loadPEMPublicKey(contents)
	if err != nil {
		return nil, err
	}

	return &pemVerifier{
		publicKey: publicKey,
		keyID:     sha256Hex(rawVerifier),
	}, nil

}

func loadPEMSigner(path string) (crypto.Signer, *pemVerifier, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read signer %q: %w", path, err)
	}

	block, _ := pem.Decode(contents)
	if block == nil {
		return nil, nil, fmt.Errorf("decode signer %q: no PEM block found", path)
	}

	var signer crypto.Signer
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		var ok bool
		signer, ok = key.(crypto.Signer)
		if !ok {
			return nil, nil, fmt.Errorf("signer %q does not implement crypto.Signer", path)
		}
	} else if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		signer = key
	} else if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		signer = key
	} else {
		return nil, nil, fmt.Errorf("unsupported private key in signer %q", path)
	}

	publicKeyPEM, err := marshalPublicKeyPEM(signer.Public())
	if err != nil {
		return nil, nil, err
	}

	verifier, err := LoadPEMVerifierBytes(publicKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	return signer, verifier, nil
}

func loadPEMPublicKey(contents []byte) (crypto.PublicKey, []byte, error) {
	block, _ := pem.Decode(contents)
	if block == nil {
		return nil, nil, fmt.Errorf("no PEM block found")
	}

	if publicKey, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		return publicKey, pem.EncodeToMemory(block), nil
	}
	if certificate, err := x509.ParseCertificate(block.Bytes); err == nil {
		return certificate.PublicKey, pem.EncodeToMemory(block), nil
	}
	if privateKey, verifier, err := loadPEMPrivateKeyAsPublic(contents); err == nil {
		return privateKey, verifier, nil
	}

	return nil, nil, fmt.Errorf("unsupported PEM verifier data")
}

func loadPEMPrivateKeyAsPublic(contents []byte) (crypto.PublicKey, []byte, error) {
	block, _ := pem.Decode(contents)
	if block == nil {
		return nil, nil, fmt.Errorf("no PEM block found")
	}

	var signer crypto.Signer
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		var ok bool
		signer, ok = key.(crypto.Signer)
		if !ok {
			return nil, nil, fmt.Errorf("private key does not implement crypto.Signer")
		}
	} else if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		signer = key
	} else if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		signer = key
	} else {
		return nil, nil, fmt.Errorf("unsupported private key data")
	}

	publicKeyPEM, err := marshalPublicKeyPEM(signer.Public())
	if err != nil {
		return nil, nil, err
	}

	publicKey, _, err := loadPEMPublicKey(publicKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	return publicKey, publicKeyPEM, nil
}

func marshalPublicKeyPEM(publicKey crypto.PublicKey) ([]byte, error) {
	encoded, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded}), nil
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
