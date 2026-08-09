// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	internalattestation "github.com/hashicorp/packer/internal/attestation"
	internalprovenance "github.com/hashicorp/packer/internal/provenance"
)

func TestVerifyAttestationCommandRun(t *testing.T) {
	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello verify"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	privateKeyPath, publicKeyPath := writeVerifierKeypair(t)
	attestationPath := writeSignedAttestation(t, privateKeyPath, artifactPath, "https://example.com/builder", "git+https://github.com/hashicorp/packer@refs/heads/main")

	meta := testMeta(t)
	command := &VerifyAttestationCommand{Meta: meta}
	ret := command.Run([]string{
		"-verifier=" + publicKeyPath,
		"-predicate-type=" + internalprovenance.SLSAProvenanceV1PredicateType,
		"-builder-id=https://example.com/builder",
		"-source-uri=git+https://github.com/hashicorp/packer@refs/heads/main",
		"-artifact=" + artifactPath,
		attestationPath,
	})
	if ret != 0 {
		fatalCommand(t, meta)
	}

	ui := meta.Ui.(*packersdk.BasicUi)
	if got := ui.Writer.(*bytes.Buffer).String(); !strings.Contains(got, "Attestation verified.") {
		t.Fatalf("expected success output, got %q", got)
	}
}

func TestVerifyAttestationCommandRejectsMismatchedBuilderID(t *testing.T) {
	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello verify"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	privateKeyPath, publicKeyPath := writeVerifierKeypair(t)
	attestationPath := writeSignedAttestation(t, privateKeyPath, artifactPath, "https://example.com/builder", "git+https://github.com/hashicorp/packer@refs/heads/main")

	meta := testMeta(t)
	command := &VerifyAttestationCommand{Meta: meta}
	ret := command.Run([]string{
		"-verifier=" + publicKeyPath,
		"-builder-id=https://example.com/other-builder",
		attestationPath,
	})
	if ret == 0 {
		t.Fatalf("expected builder mismatch to fail")
	}
}

func TestVerifyAttestationCommandRequiresBundleForRekorChecks(t *testing.T) {
	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello verify"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	privateKeyPath, publicKeyPath := writeVerifierKeypair(t)
	attestationPath := writeSignedAttestation(t, privateKeyPath, artifactPath, "https://example.com/builder", "git+https://github.com/hashicorp/packer@refs/heads/main")

	meta := testMeta(t)
	command := &VerifyAttestationCommand{Meta: meta}
	ret := command.Run([]string{
		"-verifier=" + publicKeyPath,
		"-require-rekor",
		attestationPath,
	})
	if ret == 0 {
		t.Fatalf("expected missing bundle to fail")
	}

	ui := meta.Ui.(*packersdk.BasicUi)
	if got := ui.ErrorWriter.(*bytes.Buffer).String(); !strings.Contains(got, "requires -bundle") {
		t.Fatalf("expected missing bundle error, got %q", got)
	}
}

func TestVerifyAttestationCommandRejectsTamperedPayload(t *testing.T) {
	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello verify"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	privateKeyPath, publicKeyPath := writeVerifierKeypair(t)
	attestationPath := writeSignedAttestation(t, privateKeyPath, artifactPath, "https://example.com/builder", "git+https://github.com/hashicorp/packer@refs/heads/main")

	tamperAttestationPayload(t, attestationPath)

	meta := testMeta(t)
	command := &VerifyAttestationCommand{Meta: meta}
	ret := command.Run([]string{
		"-verifier=" + publicKeyPath,
		attestationPath,
	})
	if ret == 0 {
		t.Fatalf("expected tampered payload to fail verification")
	}
}

func tamperAttestationPayload(t *testing.T, attestationPath string) {
	t.Helper()

	contents, err := os.ReadFile(attestationPath)
	if err != nil {
		t.Fatalf("read attestation: %v", err)
	}

	var envelope internalattestation.Envelope
	if err := json.Unmarshal(contents, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	payload, err := internalattestation.DecodeEnvelopePayload(envelope)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	tampered := bytes.Replace(payload, []byte("https://example.com/builder"), []byte("https://example.com/evilbuilder"), 1)
	if bytes.Equal(tampered, payload) {
		t.Fatalf("expected payload to be modified")
	}
	envelope.Payload = base64.StdEncoding.EncodeToString(tampered)

	tamperedEnvelope, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal tampered envelope: %v", err)
	}
	if err := os.WriteFile(attestationPath, tamperedEnvelope, 0600); err != nil {
		t.Fatalf("write tampered attestation: %v", err)
	}
}

func writeSignedAttestation(t *testing.T, privateKeyPath, artifactPath, builderID, sourceURI string) string {
	t.Helper()

	artifactDigest := sha256.Sum256([]byte(readFileString(t, artifactPath)))
	statement := internalprovenance.WrapInToto(
		[]internalprovenance.Subject{{
			Name: filepath.Base(artifactPath),
			Digest: internalprovenance.DigestSet{
				"sha256": hex.EncodeToString(artifactDigest[:]),
			},
		}},
		internalprovenance.SLSAProvenanceV1PredicateType,
		internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{
			BuilderID:            builderID,
			ResolvedDependencies: []internalprovenance.ResolvedDependency{{URI: sourceURI}},
		}),
	)
	payload, err := internalattestation.MarshalPayload(statement)
	if err != nil {
		t.Fatalf("marshal statement: %v", err)
	}

	signer, err := internalattestation.NewSigner(context.Background(), internalattestation.BackendConfig{Mode: internalattestation.SigningModeKey, SignerRef: privateKeyPath})
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}

	signature, err := signer.Sign(context.Background(), internalattestation.InTotoPayloadType, payload)
	if err != nil {
		t.Fatalf("sign statement: %v", err)
	}

	attestationPath := filepath.Join(t.TempDir(), "attestation.json")
	envelope, err := json.Marshal(internalattestation.NewEnvelope(internalattestation.InTotoPayloadType, payload, signature))
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := os.WriteFile(attestationPath, envelope, 0600); err != nil {
		t.Fatalf("write attestation: %v", err)
	}

	return attestationPath
}

func writeVerifierKeypair(t *testing.T) (string, string) {
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

func readFileString(t *testing.T, path string) string {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}

	return string(contents)
}
