// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	internalprovenance "github.com/hashicorp/packer/internal/provenance"
	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	sigstorebundle "github.com/sigstore/sigstore-go/pkg/bundle"
	fulciocertificate "github.com/sigstore/sigstore-go/pkg/fulcio/certificate"
	sigstoreroot "github.com/sigstore/sigstore-go/pkg/root"
	sigstoregosign "github.com/sigstore/sigstore-go/pkg/sign"
	sigstoresignature "github.com/sigstore/sigstore/pkg/signature"
	sigstorekms "github.com/sigstore/sigstore/pkg/signature/kms"
)

func TestKMSSignerAndVerifier(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	localSignerVerifier, err := sigstoresignature.LoadDefaultSignerVerifier(privateKey)
	if err != nil {
		t.Fatalf("load local signer verifier: %v", err)
	}

	originalFactory := newKMSSignerVerifier
	newKMSSignerVerifier = func(ctx context.Context, keyResourceID string) (sigstorekms.SignerVerifier, error) {
		if got, want := keyResourceID, "awskms://alias/example"; got != want {
			t.Fatalf("unexpected KMS key resource %q, want %q", got, want)
		}
		return &fakeKMSSignerVerifier{
			SignerVerifier: localSignerVerifier,
			publicKey:      privateKey.Public(),
		}, nil
	}
	t.Cleanup(func() {
		newKMSSignerVerifier = originalFactory
	})

	signer, err := NewSigner(context.Background(), BackendConfig{Mode: SigningModeKMS, SignerRef: "awskms://alias/example"})
	if err != nil {
		t.Fatalf("create KMS signer: %v", err)
	}

	verifier, err := NewVerifier(context.Background(), BackendConfig{Mode: SigningModeKMS}, signer)
	if err != nil {
		t.Fatalf("create KMS verifier: %v", err)
	}

	payload := []byte(`{"hello":"kms"}`)
	signature, err := signer.Sign(context.Background(), InTotoPayloadType, payload)
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}
	if signature.KeyID == "" {
		t.Fatalf("expected KMS signature key ID to be populated")
	}

	envelope := NewEnvelope(InTotoPayloadType, payload, signature)
	if err := VerifyEnvelope(context.Background(), envelope, verifier); err != nil {
		t.Fatalf("verify envelope: %v", err)
	}
}

func TestKeylessSignerAndVerifier(t *testing.T) {
	originalKeypairFactory := newKeylessEphemeralKeypair
	originalFulcioFactory := newKeylessFulcio
	originalTrustedMaterialLoader := loadKeylessTrustedMaterial
	originalCertificateVerifier := verifyKeylessCertificate

	newKeylessEphemeralKeypair = func() (sigstoregosign.Keypair, error) {
		return sigstoregosign.NewEphemeralKeypair(nil)
	}
	newKeylessFulcio = func(baseURL string) sigstoregosign.CertificateProvider {
		if got, want := baseURL, "https://fulcio.example.test"; got != want {
			t.Fatalf("unexpected Fulcio URL %q, want %q", got, want)
		}
		return fakeCertificateProvider{t: t, wantToken: "test-oidc-token"}
	}
	loadKeylessTrustedMaterial = func(cfg BackendConfig) (sigstoreroot.TrustedMaterial, error) {
		if got, want := cfg.TrustedRootPath, "/tmp/test-root.json"; got != want {
			t.Fatalf("unexpected trusted root path %q, want %q", got, want)
		}
		return nil, nil
	}
	verifyKeylessCertificate = func(certificate *x509.Certificate, trustedMaterial sigstoreroot.TrustedMaterial, expectedIdentity, expectedOIDCIssuer string) error {
		summary, err := fulciocertificate.SummarizeCertificate(certificate)
		if err != nil {
			t.Fatalf("summarize certificate: %v", err)
		}
		if got, want := summary.SubjectAlternativeName, "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"; got != want {
			t.Fatalf("unexpected certificate SAN %q, want %q", got, want)
		}
		if got, want := summary.Issuer, "https://token.actions.githubusercontent.com"; got != want {
			t.Fatalf("unexpected certificate OIDC issuer %q, want %q", got, want)
		}
		if got, want := expectedIdentity, "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"; got != want {
			t.Fatalf("unexpected expected identity %q, want %q", got, want)
		}
		if got, want := expectedOIDCIssuer, "https://token.actions.githubusercontent.com"; got != want {
			t.Fatalf("unexpected expected OIDC issuer %q, want %q", got, want)
		}
		return nil
	}
	t.Cleanup(func() {
		newKeylessEphemeralKeypair = originalKeypairFactory
		newKeylessFulcio = originalFulcioFactory
		loadKeylessTrustedMaterial = originalTrustedMaterialLoader
		verifyKeylessCertificate = originalCertificateVerifier
	})

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode:      SigningModeKeyless,
		Env:       map[string]string{"SIGSTORE_ID_TOKEN": "test-oidc-token"},
		FulcioURL: "https://fulcio.example.test",
	})
	if err != nil {
		t.Fatalf("create keyless signer: %v", err)
	}

	verifier, err := NewVerifier(context.Background(), BackendConfig{
		Mode:              SigningModeKeyless,
		TrustedRootPath:   "/tmp/test-root.json",
		KeylessIdentity:   "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main",
		KeylessOIDCIssuer: "https://token.actions.githubusercontent.com",
	}, signer)
	if err != nil {
		t.Fatalf("create keyless verifier: %v", err)
	}

	payload := []byte(`{"hello":"keyless"}`)
	signature, err := signer.Sign(context.Background(), InTotoPayloadType, payload)
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}
	if len(signature.CertPEM) == 0 {
		t.Fatalf("expected keyless signature to embed a certificate")
	}

	envelope := NewEnvelope(InTotoPayloadType, payload, signature)
	if err := VerifyEnvelope(context.Background(), envelope, verifier); err != nil {
		t.Fatalf("verify envelope: %v", err)
	}
}

func TestBuildBundleForKeylessSigner(t *testing.T) {
	originalKeypairFactory := newKeylessEphemeralKeypair
	originalFulcioFactory := newKeylessFulcio
	originalBundleFactory := newKeylessBundle
	originalRekorFactory := newKeylessRekor
	originalTrustedMaterialLoader := loadKeylessTrustedMaterial

	newKeylessEphemeralKeypair = func() (sigstoregosign.Keypair, error) {
		return sigstoregosign.NewEphemeralKeypair(nil)
	}
	newKeylessFulcio = func(string) sigstoregosign.CertificateProvider {
		return fakeCertificateProvider{t: t, wantToken: "test-oidc-token"}
	}
	loadKeylessTrustedMaterial = func(cfg BackendConfig) (sigstoreroot.TrustedMaterial, error) {
		if got, want := cfg.TrustedRootPath, "/tmp/test-root.json"; got != want {
			t.Fatalf("unexpected trusted root path %q, want %q", got, want)
		}
		return nil, nil
	}
	rekorCalled := false
	newKeylessRekor = func(baseURL string) sigstoregosign.Transparency {
		if got, want := baseURL, "https://rekor.example.test"; got != want {
			t.Fatalf("unexpected Rekor URL %q, want %q", got, want)
		}
		rekorCalled = true
		return fakeTransparency{}
	}
	newKeylessBundle = func(content sigstoregosign.Content, keypair sigstoregosign.Keypair, opts sigstoregosign.BundleOptions) (*protobundle.Bundle, error) {
		if len(opts.TransparencyLogs) != 1 {
			t.Fatalf("expected one transparency log, got %d", len(opts.TransparencyLogs))
		}
		return sigstoregosign.Bundle(content, keypair, sigstoregosign.BundleOptions{
			CertificateProvider: opts.CertificateProvider,
			Context:             opts.Context,
			TrustedRoot:         opts.TrustedRoot,
		})
	}
	t.Cleanup(func() {
		newKeylessEphemeralKeypair = originalKeypairFactory
		newKeylessFulcio = originalFulcioFactory
		newKeylessBundle = originalBundleFactory
		newKeylessRekor = originalRekorFactory
		loadKeylessTrustedMaterial = originalTrustedMaterialLoader
	})

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode:      SigningModeKeyless,
		Env:       map[string]string{"SIGSTORE_ID_TOKEN": "test-oidc-token"},
		FulcioURL: "https://fulcio.example.test",
	})
	if err != nil {
		t.Fatalf("create keyless signer: %v", err)
	}

	envelope, bundleJSON, err := BuildBundleForSigner(context.Background(), signer, BackendConfig{
		Mode:            SigningModeKeyless,
		UploadTlog:      true,
		RekorURL:        "https://rekor.example.test",
		TrustedRootPath: "/tmp/test-root.json",
	}, InTotoPayloadType, []byte(`{"hello":"bundle"}`))
	if err != nil {
		t.Fatalf("build bundle: %v", err)
	}
	if !rekorCalled {
		t.Fatalf("expected Rekor constructor to be used")
	}
	if got, want := envelope.PayloadType, InTotoPayloadType; got != want {
		t.Fatalf("unexpected envelope payload type %q, want %q", got, want)
	}
	if len(envelope.Signatures) != 1 || envelope.Signatures[0].Cert == "" {
		t.Fatalf("expected bundled envelope to include one certificate-backed signature")
	}

	bundlePath := filepath.Join(t.TempDir(), "bundle.json")
	if err := os.WriteFile(bundlePath, bundleJSON, 0600); err != nil {
		t.Fatalf("write bundle: %v", err)
	}
	if _, err := sigstorebundle.LoadJSONFromPath(bundlePath); err != nil {
		t.Fatalf("load bundle: %v", err)
	}
}

func TestKeylessBundleAndRekorIntegration(t *testing.T) {
	if os.Getenv("PACKER_ACC") == "" {
		t.Skip("acceptance-style keyless integration test skipped unless PACKER_ACC is set")
	}

	env := currentProcessEnv()
	if _, err := resolveAmbientIDToken(env); err != nil {
		t.Skipf("keyless integration test skipped without an ambient OIDC token: %v", err)
	}

	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello live keyless"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	digest, err := sha256File(artifactPath)
	if err != nil {
		t.Fatalf("hash artifact: %v", err)
	}

	statement := internalprovenance.WrapInToto(
		[]internalprovenance.Subject{{Name: filepath.Base(artifactPath), Digest: internalprovenance.DigestSet{"sha256": digest}}},
		internalprovenance.SLSAProvenanceV1PredicateType,
		internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{}),
	)
	payload, err := MarshalPayload(statement)
	if err != nil {
		t.Fatalf("marshal statement: %v", err)
	}

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode: SigningModeKeyless,
		Env:  env,
	})
	if err != nil {
		t.Fatalf("create live keyless signer: %v", err)
	}

	envelope, bundleJSON, err := BuildBundleForSigner(context.Background(), signer, BackendConfig{
		Mode:       SigningModeKeyless,
		Env:        env,
		UploadTlog: true,
	}, InTotoPayloadType, payload)
	if err != nil {
		t.Fatalf("build live Sigstore bundle: %v", err)
	}

	certificate, err := certificateFromEnvelope(envelope)
	if err != nil {
		t.Fatalf("extract certificate from live envelope: %v", err)
	}
	summary, err := fulciocertificate.SummarizeCertificate(certificate)
	if err != nil {
		t.Fatalf("summarize live certificate: %v", err)
	}
	if summary.SubjectAlternativeName == "" || summary.Issuer == "" {
		t.Fatalf("expected live certificate to contain SAN and issuer, got SAN=%q issuer=%q", summary.SubjectAlternativeName, summary.Issuer)
	}

	envelopePath := filepath.Join(t.TempDir(), "attestation.json")
	if err := os.WriteFile(envelopePath, mustJSONMarshal(t, envelope), 0600); err != nil {
		t.Fatalf("write live envelope: %v", err)
	}
	bundlePath := filepath.Join(t.TempDir(), "attestation.sigstore.json")
	if err := os.WriteFile(bundlePath, bundleJSON, 0600); err != nil {
		t.Fatalf("write live bundle: %v", err)
	}
	if _, err := sigstorebundle.LoadJSONFromPath(bundlePath); err != nil {
		t.Fatalf("load live bundle: %v", err)
	}

	verifiedStatement, err := VerifyAttestationFile(context.Background(), envelopePath, BackendConfig{
		Mode:              SigningModeKeyless,
		KeylessIdentity:   summary.SubjectAlternativeName,
		KeylessOIDCIssuer: summary.Issuer,
	}, VerificationPolicy{
		PredicateType:          internalprovenance.SLSAProvenanceV1PredicateType,
		ArtifactPath:           artifactPath,
		SigstoreBundlePath:     bundlePath,
		RequireTransparencyLog: true,
	})
	if err != nil {
		t.Fatalf("verify live attestation with Rekor bundle: %v", err)
	}
	if got, want := verifiedStatement.PredicateType, internalprovenance.SLSAProvenanceV1PredicateType; got != want {
		t.Fatalf("unexpected predicate type %q, want %q", got, want)
	}
}

func TestKeylessSignerRequiresAmbientToken(t *testing.T) {
	_, err := NewSigner(context.Background(), BackendConfig{Mode: SigningModeKeyless, Env: map[string]string{}})
	if err == nil {
		t.Fatalf("expected keyless signer creation to fail without an ambient token")
	}
	if !strings.Contains(err.Error(), "ambient OIDC token") {
		t.Fatalf("unexpected keyless token error: %v", err)
	}
}

func TestKeylessVerifierRequiresIdentityPolicy(t *testing.T) {
	originalKeypairFactory := newKeylessEphemeralKeypair
	originalFulcioFactory := newKeylessFulcio
	originalTrustedMaterialLoader := loadKeylessTrustedMaterial

	newKeylessEphemeralKeypair = func() (sigstoregosign.Keypair, error) {
		return sigstoregosign.NewEphemeralKeypair(nil)
	}
	newKeylessFulcio = func(string) sigstoregosign.CertificateProvider {
		return fakeCertificateProvider{t: t, wantToken: "test-oidc-token"}
	}
	loadKeylessTrustedMaterial = func(cfg BackendConfig) (sigstoreroot.TrustedMaterial, error) {
		return nil, nil
	}
	t.Cleanup(func() {
		newKeylessEphemeralKeypair = originalKeypairFactory
		newKeylessFulcio = originalFulcioFactory
		loadKeylessTrustedMaterial = originalTrustedMaterialLoader
	})

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode:      SigningModeKeyless,
		Env:       map[string]string{"SIGSTORE_ID_TOKEN": "test-oidc-token"},
		FulcioURL: "https://fulcio.example.test",
	})
	if err != nil {
		t.Fatalf("create keyless signer: %v", err)
	}

	_, err = NewVerifier(context.Background(), BackendConfig{Mode: SigningModeKeyless}, signer)
	if err == nil {
		t.Fatalf("expected keyless verifier creation to fail without identity policy")
	}
	if !strings.Contains(err.Error(), "keyless_identity") {
		t.Fatalf("unexpected keyless verifier error: %v", err)
	}
}

func TestVerifyAttestationFileWithKeylessEnvelope(t *testing.T) {
	originalKeypairFactory := newKeylessEphemeralKeypair
	originalFulcioFactory := newKeylessFulcio
	originalTrustedMaterialLoader := loadKeylessTrustedMaterial
	originalCertificateVerifier := verifyKeylessCertificate

	newKeylessEphemeralKeypair = func() (sigstoregosign.Keypair, error) {
		return sigstoregosign.NewEphemeralKeypair(nil)
	}
	newKeylessFulcio = func(string) sigstoregosign.CertificateProvider {
		return fakeCertificateProvider{t: t, wantToken: "test-oidc-token"}
	}
	loadKeylessTrustedMaterial = func(cfg BackendConfig) (sigstoreroot.TrustedMaterial, error) {
		if got, want := cfg.TrustedRootPath, "/tmp/test-root.json"; got != want {
			t.Fatalf("unexpected trusted root path %q, want %q", got, want)
		}
		return nil, nil
	}
	verifyKeylessCertificate = func(certificate *x509.Certificate, trustedMaterial sigstoreroot.TrustedMaterial, expectedIdentity, expectedOIDCIssuer string) error {
		summary, err := fulciocertificate.SummarizeCertificate(certificate)
		if err != nil {
			t.Fatalf("summarize certificate: %v", err)
		}
		if got, want := summary.SubjectAlternativeName, "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"; got != want {
			t.Fatalf("unexpected certificate SAN %q, want %q", got, want)
		}
		if got, want := expectedIdentity, "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"; got != want {
			t.Fatalf("unexpected expected identity %q, want %q", got, want)
		}
		if got, want := expectedOIDCIssuer, "https://token.actions.githubusercontent.com"; got != want {
			t.Fatalf("unexpected expected OIDC issuer %q, want %q", got, want)
		}
		return nil
	}
	t.Cleanup(func() {
		newKeylessEphemeralKeypair = originalKeypairFactory
		newKeylessFulcio = originalFulcioFactory
		loadKeylessTrustedMaterial = originalTrustedMaterialLoader
		verifyKeylessCertificate = originalCertificateVerifier
	})

	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello keyless"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	digest, err := sha256File(artifactPath)
	if err != nil {
		t.Fatalf("hash artifact: %v", err)
	}

	statement := internalprovenance.WrapInToto(
		[]internalprovenance.Subject{{Name: filepath.Base(artifactPath), Digest: internalprovenance.DigestSet{"sha256": digest}}},
		internalprovenance.SLSAProvenanceV1PredicateType,
		internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{
			BuilderID:            "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main",
			ResolvedDependencies: []internalprovenance.ResolvedDependency{{URI: "git+https://github.com/hashicorp/packer@refs/heads/main"}},
		}),
	)
	payload, err := MarshalPayload(statement)
	if err != nil {
		t.Fatalf("marshal statement: %v", err)
	}

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode:      SigningModeKeyless,
		Env:       map[string]string{"SIGSTORE_ID_TOKEN": "test-oidc-token"},
		FulcioURL: "https://fulcio.example.test",
	})
	if err != nil {
		t.Fatalf("create keyless signer: %v", err)
	}

	signature, err := signer.Sign(context.Background(), InTotoPayloadType, payload)
	if err != nil {
		t.Fatalf("sign statement: %v", err)
	}

	envelopePath := filepath.Join(t.TempDir(), "attestation.json")
	envelopeJSON, err := json.Marshal(NewEnvelope(InTotoPayloadType, payload, signature))
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := os.WriteFile(envelopePath, envelopeJSON, 0600); err != nil {
		t.Fatalf("write envelope: %v", err)
	}

	verifiedStatement, err := VerifyAttestationFile(context.Background(), envelopePath, BackendConfig{
		Mode:              SigningModeKeyless,
		TrustedRootPath:   "/tmp/test-root.json",
		KeylessIdentity:   "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main",
		KeylessOIDCIssuer: "https://token.actions.githubusercontent.com",
	}, VerificationPolicy{
		PredicateType: internalprovenance.SLSAProvenanceV1PredicateType,
		BuilderID:     "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main",
		SourceURI:     "git+https://github.com/hashicorp/packer@refs/heads/main",
		ArtifactPath:  artifactPath,
	})
	if err != nil {
		t.Fatalf("verify attestation file: %v", err)
	}
	if got, want := verifiedStatement.PredicateType, internalprovenance.SLSAProvenanceV1PredicateType; got != want {
		t.Fatalf("unexpected predicate type %q, want %q", got, want)
	}
}

func TestVerifyAttestationFileRequiresBundleForTransparencyChecks(t *testing.T) {
	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello keyless"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	digest, err := sha256File(artifactPath)
	if err != nil {
		t.Fatalf("hash artifact: %v", err)
	}

	statement := internalprovenance.WrapInToto(
		[]internalprovenance.Subject{{Name: filepath.Base(artifactPath), Digest: internalprovenance.DigestSet{"sha256": digest}}},
		internalprovenance.SLSAProvenanceV1PredicateType,
		internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{}),
	)
	payload, err := MarshalPayload(statement)
	if err != nil {
		t.Fatalf("marshal statement: %v", err)
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	verifier, err := newSigstoreVerifierFromPublicKey(privateKey.Public())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	sig, err := privateKey.Sign(rand.Reader, PreAuthEncode(InTotoPayloadType, payload), crypto.SHA256)
	if err != nil {
		t.Fatalf("sign payload: %v", err)
	}

	envelopePath := filepath.Join(t.TempDir(), "attestation.json")
	envelopeJSON, err := json.Marshal(NewEnvelope(InTotoPayloadType, payload, Signature{Sig: sig, CertPEM: []byte("-----BEGIN CERTIFICATE-----\nZmFrZQ==\n-----END CERTIFICATE-----\n")}))
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := os.WriteFile(envelopePath, envelopeJSON, 0600); err != nil {
		t.Fatalf("write envelope: %v", err)
	}

	_ = verifier
	_, err = VerifyAttestationFile(context.Background(), envelopePath, BackendConfig{
		Mode:              SigningModeKeyless,
		KeylessIdentity:   "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main",
		KeylessOIDCIssuer: "https://token.actions.githubusercontent.com",
	}, VerificationPolicy{RequireTransparencyLog: true})
	if err == nil || !strings.Contains(err.Error(), "requires -bundle") {
		t.Fatalf("expected missing bundle error, got %v", err)
	}
}

func TestVerifyAttestationFileWithBundleRequirements(t *testing.T) {
	originalKeypairFactory := newKeylessEphemeralKeypair
	originalFulcioFactory := newKeylessFulcio
	originalTrustedMaterialLoader := loadKeylessTrustedMaterial
	originalCertificateVerifier := verifyKeylessCertificate
	originalBundleVerifier := verifySigstoreBundleEvidence

	newKeylessEphemeralKeypair = func() (sigstoregosign.Keypair, error) {
		return sigstoregosign.NewEphemeralKeypair(nil)
	}
	newKeylessFulcio = func(string) sigstoregosign.CertificateProvider {
		return fakeCertificateProvider{t: t, wantToken: "test-oidc-token"}
	}
	loadKeylessTrustedMaterial = func(cfg BackendConfig) (sigstoreroot.TrustedMaterial, error) {
		return nil, nil
	}
	verifyKeylessCertificate = func(*x509.Certificate, sigstoreroot.TrustedMaterial, string, string) error {
		return nil
	}
	called := false
	verifySigstoreBundleEvidence = func(envelope Envelope, cfg BackendConfig, policy VerificationPolicy) error {
		called = true
		if got, want := policy.SigstoreBundlePath, "bundle.json"; got != want {
			t.Fatalf("unexpected bundle path %q, want %q", got, want)
		}
		if !policy.RequireTransparencyLog || !policy.RequireObserverTimestamp {
			t.Fatalf("expected Rekor and timestamp requirements to be set")
		}
		if len(envelope.Signatures) != 1 {
			t.Fatalf("unexpected signature count %d", len(envelope.Signatures))
		}
		if got, want := cfg.KeylessIdentity, "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"; got != want {
			t.Fatalf("unexpected keyless identity %q, want %q", got, want)
		}
		return nil
	}
	t.Cleanup(func() {
		newKeylessEphemeralKeypair = originalKeypairFactory
		newKeylessFulcio = originalFulcioFactory
		loadKeylessTrustedMaterial = originalTrustedMaterialLoader
		verifyKeylessCertificate = originalCertificateVerifier
		verifySigstoreBundleEvidence = originalBundleVerifier
	})

	artifactPath := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("hello keyless"), 0600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	digest, err := sha256File(artifactPath)
	if err != nil {
		t.Fatalf("hash artifact: %v", err)
	}

	statement := internalprovenance.WrapInToto(
		[]internalprovenance.Subject{{Name: filepath.Base(artifactPath), Digest: internalprovenance.DigestSet{"sha256": digest}}},
		internalprovenance.SLSAProvenanceV1PredicateType,
		internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{}),
	)
	payload, err := MarshalPayload(statement)
	if err != nil {
		t.Fatalf("marshal statement: %v", err)
	}

	signer, err := NewSigner(context.Background(), BackendConfig{
		Mode:      SigningModeKeyless,
		Env:       map[string]string{"SIGSTORE_ID_TOKEN": "test-oidc-token"},
		FulcioURL: "https://fulcio.example.test",
	})
	if err != nil {
		t.Fatalf("create keyless signer: %v", err)
	}

	signature, err := signer.Sign(context.Background(), InTotoPayloadType, payload)
	if err != nil {
		t.Fatalf("sign statement: %v", err)
	}

	envelopePath := filepath.Join(t.TempDir(), "attestation.json")
	envelopeJSON, err := json.Marshal(NewEnvelope(InTotoPayloadType, payload, signature))
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	if err := os.WriteFile(envelopePath, envelopeJSON, 0600); err != nil {
		t.Fatalf("write envelope: %v", err)
	}

	_, err = VerifyAttestationFile(context.Background(), envelopePath, BackendConfig{
		Mode:              SigningModeKeyless,
		KeylessIdentity:   "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main",
		KeylessOIDCIssuer: "https://token.actions.githubusercontent.com",
	}, VerificationPolicy{
		SigstoreBundlePath:       "bundle.json",
		RequireTransparencyLog:   true,
		RequireObserverTimestamp: true,
	})
	if err != nil {
		t.Fatalf("verify attestation with bundle requirements: %v", err)
	}
	if !called {
		t.Fatalf("expected Sigstore bundle verifier to be called")
	}
	_ = sigstorebundle.Bundle{}
}

type fakeTransparency struct{}

func (fakeTransparency) GetTransparencyLogEntry(context.Context, []byte, *protobundle.Bundle) error {
	return nil
}

func currentProcessEnv() map[string]string {
	env := make(map[string]string)
	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			continue
		}
		env[parts[0]] = parts[1]
	}

	return env
}

func mustJSONMarshal(t *testing.T, value interface{}) []byte {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}

	return encoded
}

type fakeKMSSignerVerifier struct {
	sigstoresignature.SignerVerifier
	publicKey crypto.PublicKey
}

func (f *fakeKMSSignerVerifier) CreateKey(context.Context, string) (crypto.PublicKey, error) {
	return f.publicKey, nil
}

func (f *fakeKMSSignerVerifier) CryptoSigner(context.Context, func(error)) (crypto.Signer, crypto.SignerOpts, error) {
	return nil, nil, fmt.Errorf("not implemented in tests")
}

func (f *fakeKMSSignerVerifier) SupportedAlgorithms() []string {
	return []string{"ecdsa-p256-sha256"}
}

func (f *fakeKMSSignerVerifier) DefaultAlgorithm() string {
	return "ecdsa-p256-sha256"
}

type fakeCertificateProvider struct {
	t         *testing.T
	wantToken string
}

func (f fakeCertificateProvider) GetCertificate(_ context.Context, keypair sigstoregosign.Keypair, opts *sigstoregosign.CertificateProviderOptions) ([]byte, error) {
	f.t.Helper()
	if opts == nil {
		f.t.Fatalf("expected certificate options to be provided")
	}
	if got, want := opts.IDToken, f.wantToken; got != want {
		f.t.Fatalf("unexpected Fulcio ID token %q, want %q", got, want)
	}
	return createCertificateDER(f.t, keypair.GetPublicKey()), nil
}

func createCertificateDER(t *testing.T, publicKey crypto.PublicKey) []byte {
	t.Helper()

	issuerKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate issuer key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "packer-keyless-test",
		},
		URIs:                  []*url.URL{mustParseURL(t, "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main")},
		ExtraExtensions:       []pkix.Extension{buildIssuerExtension(t, "https://token.actions.githubusercontent.com")},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, publicKey, issuerKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	return der
}

func buildIssuerExtension(t *testing.T, issuer string) pkix.Extension {
	t.Helper()

	value, err := asn1.MarshalWithParams(issuer, "utf8")
	if err != nil {
		t.Fatalf("marshal issuer extension: %v", err)
	}

	return pkix.Extension{Id: fulciocertificate.OIDIssuerV2, Value: value}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url %q: %v", rawURL, err)
	}

	return parsedURL
}
