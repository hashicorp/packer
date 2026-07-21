// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestPEMSignerAndVerifier(t *testing.T) {
	privateKeyPath, publicKeyPath := writeECDSAKeypair(t)

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode:      SigningModeKey,
		SignerRef: privateKeyPath,
	})
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}

	verifier, err := NewVerifier(context.Background(), BackendConfig{
		Mode:        SigningModeKey,
		VerifierRef: publicKeyPath,
	}, signer)
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	payload := []byte(`{"hello":"world"}`)
	signature, err := signer.Sign(context.Background(), InTotoPayloadType, payload)
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}

	envelope := NewEnvelope(InTotoPayloadType, payload, signature)
	if err := VerifyEnvelope(context.Background(), envelope, verifier); err != nil {
		t.Fatalf("verify envelope: %v", err)
	}
}

func TestVerifierOverrideMismatchFails(t *testing.T) {
	privateKeyPath, _ := writeECDSAKeypair(t)
	_, mismatchedPublicKeyPath := writeECDSAKeypair(t)

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode:      SigningModeKey,
		SignerRef: privateKeyPath,
	})
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}

	verifier, err := NewVerifier(context.Background(), BackendConfig{
		Mode:        SigningModeKey,
		VerifierRef: mismatchedPublicKeyPath,
	}, signer)
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}

	signature, err := signer.Sign(context.Background(), InTotoPayloadType, []byte(`{"hello":"world"}`))
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}

	envelope := NewEnvelope(InTotoPayloadType, []byte(`{"hello":"world"}`), signature)
	if err := VerifyEnvelope(context.Background(), envelope, verifier); err == nil {
		t.Fatalf("expected verifier mismatch to fail")
	}
}

func writeECDSAKeypair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	privateKeyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	dir := t.TempDir()
	privateKeyPath := filepath.Join(dir, "signer.pem")
	publicKeyPath := filepath.Join(dir, "verifier.pem")

	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyDER}), 0600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	if err := os.WriteFile(publicKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER}), 0600); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	return privateKeyPath, publicKeyPath
}
