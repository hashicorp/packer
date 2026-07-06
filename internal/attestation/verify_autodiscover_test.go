// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	internalprovenance "github.com/hashicorp/packer/internal/provenance"
)

func TestDefaultSigstoreBundlePath(t *testing.T) {
	cases := map[string]string{
		"artifact.provenance.json": "artifact.provenance.sigstore.json",
		"dir/attestation.json":     "dir/attestation.sigstore.json",
		"attestation":              "attestation.sigstore.json",
	}

	for input, want := range cases {
		if got := defaultSigstoreBundlePath(input); got != want {
			t.Fatalf("defaultSigstoreBundlePath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestResolveSigstoreBundlePath(t *testing.T) {
	dir := t.TempDir()
	attestationPath := filepath.Join(dir, "attestation.json")

	// Explicit path always wins.
	if got := resolveSigstoreBundlePath(VerificationPolicy{SigstoreBundlePath: "explicit.json"}, attestationPath); got != "explicit.json" {
		t.Fatalf("explicit bundle path not honored, got %q", got)
	}

	// No sidecar on disk means no auto-discovered path.
	if got := resolveSigstoreBundlePath(VerificationPolicy{}, attestationPath); got != "" {
		t.Fatalf("expected empty path when sidecar is absent, got %q", got)
	}

	// A sidecar next to the attestation is auto-discovered.
	sidecarPath := defaultSigstoreBundlePath(attestationPath)
	if err := os.WriteFile(sidecarPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
	if got := resolveSigstoreBundlePath(VerificationPolicy{}, attestationPath); got != sidecarPath {
		t.Fatalf("expected auto-discovered sidecar %q, got %q", sidecarPath, got)
	}
}

func TestVerifyAttestationFileAutoDiscoversKeylessBundle(t *testing.T) {
	originalAnchor := bundleAnchorsSigningTime
	originalBundleVerifier := verifySigstoreBundleEvidence
	t.Cleanup(func() {
		bundleAnchorsSigningTime = originalAnchor
		verifySigstoreBundleEvidence = originalBundleVerifier
	})

	dir := t.TempDir()
	attestationPath := filepath.Join(dir, "attestation.json")
	sidecarPath := defaultSigstoreBundlePath(attestationPath)

	statement := internalprovenance.WrapInToto(
		[]internalprovenance.Subject{{Name: "artifact.txt", Digest: internalprovenance.DigestSet{"sha256": "abc"}}},
		internalprovenance.SLSAProvenanceV1PredicateType,
		internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{}),
	)
	payload, err := MarshalPayload(statement)
	if err != nil {
		t.Fatalf("marshal statement: %v", err)
	}

	envelope := NewEnvelope(InTotoPayloadType, payload, Signature{
		Sig:     []byte("signature"),
		CertPEM: []byte("-----BEGIN CERTIFICATE-----\nZmFrZQ==\n-----END CERTIFICATE-----\n"),
	})
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := os.WriteFile(attestationPath, envelopeJSON, 0600); err != nil {
		t.Fatalf("write attestation: %v", err)
	}
	if err := os.WriteFile(sidecarPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("write sidecar bundle: %v", err)
	}

	bundleAnchorsSigningTime = func(path string) bool {
		if path != sidecarPath {
			t.Fatalf("unexpected bundle path passed to anchor check %q, want %q", path, sidecarPath)
		}
		return true
	}

	called := false
	verifySigstoreBundleEvidence = func(_ Envelope, _ BackendConfig, policy VerificationPolicy) error {
		called = true
		if policy.SigstoreBundlePath != sidecarPath {
			t.Fatalf("bundle verification used path %q, want auto-discovered %q", policy.SigstoreBundlePath, sidecarPath)
		}
		return nil
	}

	if _, err := VerifyAttestationFile(context.Background(), attestationPath, BackendConfig{
		Mode:              SigningModeKeyless,
		KeylessIdentity:   "identity",
		KeylessOIDCIssuer: "issuer",
	}, VerificationPolicy{}); err != nil {
		t.Fatalf("verify attestation file: %v", err)
	}

	if !called {
		t.Fatalf("expected auto-discovered bundle verification to run")
	}
}

func TestVerifyAttestationFileSkipsBundleWithoutTimeAnchor(t *testing.T) {
	originalAnchor := bundleAnchorsSigningTime
	originalBundleVerifier := verifySigstoreBundleEvidence
	t.Cleanup(func() {
		bundleAnchorsSigningTime = originalAnchor
		verifySigstoreBundleEvidence = originalBundleVerifier
	})

	dir := t.TempDir()
	attestationPath := filepath.Join(dir, "attestation.json")
	sidecarPath := defaultSigstoreBundlePath(attestationPath)

	statement := internalprovenance.WrapInToto(
		[]internalprovenance.Subject{{Name: "artifact.txt", Digest: internalprovenance.DigestSet{"sha256": "abc"}}},
		internalprovenance.SLSAProvenanceV1PredicateType,
		internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{}),
	)
	payload, err := MarshalPayload(statement)
	if err != nil {
		t.Fatalf("marshal statement: %v", err)
	}

	envelope := NewEnvelope(InTotoPayloadType, payload, Signature{
		Sig:     []byte("signature"),
		CertPEM: []byte("-----BEGIN CERTIFICATE-----\nZmFrZQ==\n-----END CERTIFICATE-----\n"),
	})
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := os.WriteFile(attestationPath, envelopeJSON, 0600); err != nil {
		t.Fatalf("write attestation: %v", err)
	}
	if err := os.WriteFile(sidecarPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("write sidecar bundle: %v", err)
	}

	// The sidecar exists but carries no trusted time source, so bundle-based
	// verification must not be used; the legacy certificate path runs instead.
	bundleAnchorsSigningTime = func(string) bool { return false }
	verifySigstoreBundleEvidence = func(Envelope, BackendConfig, VerificationPolicy) error {
		t.Fatalf("bundle verification must not run when the bundle lacks a time anchor")
		return nil
	}

	// The fallback path attempts certificate-based verification with a fake
	// certificate, which fails; the important assertion is that bundle
	// verification was not selected.
	if _, err := VerifyAttestationFile(context.Background(), attestationPath, BackendConfig{
		Mode:              SigningModeKeyless,
		KeylessIdentity:   "identity",
		KeylessOIDCIssuer: "issuer",
	}, VerificationPolicy{}); err == nil {
		t.Fatalf("expected fallback certificate verification to fail on the fake certificate")
	}
}
