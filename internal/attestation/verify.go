// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	internalprovenance "github.com/hashicorp/packer/internal/provenance"
	sigstorebundle "github.com/sigstore/sigstore-go/pkg/bundle"
	sigstoreverify "github.com/sigstore/sigstore-go/pkg/verify"
)

type VerificationPolicy struct {
	PredicateType            string
	BuilderID                string
	SourceURI                string
	ArtifactPath             string
	SigstoreBundlePath       string
	RequireTransparencyLog   bool
	RequireObserverTimestamp bool
}

var loadSigstoreBundle = sigstorebundle.LoadJSONFromPath

var newSigstoreBundleVerifier = sigstoreverify.NewVerifier

var verifySigstoreBundleEvidence = func(envelope Envelope, cfg BackendConfig, policy VerificationPolicy) error {
	return verifySigstoreBundleEvidenceImpl(envelope, cfg, policy)
}

func VerifyAttestationFile(ctx context.Context, path string, cfg BackendConfig, policy VerificationPolicy) (*internalprovenance.Statement, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read attestation %q: %w", path, err)
	}

	var envelope Envelope
	if err := json.Unmarshal(contents, &envelope); err != nil {
		return nil, fmt.Errorf("decode attestation envelope %q: %w", path, err)
	}

	if envelope.PayloadType != InTotoPayloadType {
		return nil, fmt.Errorf("attestation %q has unexpected payloadType %q (want %q)",
			path, envelope.PayloadType, InTotoPayloadType)
	}

	if err := verifyEnvelopeSignature(ctx, path, cfg, policy, envelope); err != nil {
		return nil, err
	}

	payload, err := DecodeEnvelopePayload(envelope)
	if err != nil {
		return nil, err
	}

	statement, err := verifyPolicy(payload, policy)
	if err != nil {
		return nil, err
	}

	return statement, nil
}

func verifyEnvelopeSignature(ctx context.Context, path string, cfg BackendConfig, policy VerificationPolicy, envelope Envelope) error {
	// An explicit Rekor or timestamp policy always uses bundle-based verification.
	if requiresSigstoreBundle(policy) {
		return verifySigstoreBundleEvidence(envelope, cfg, policy)
	}

	// Keyless attestations are signed with a short-lived Fulcio certificate that
	// appears expired against wall-clock time moments after signing. When a
	// Sigstore bundle carrying transparency-log or timestamp evidence is
	// available, prefer it so the certificate is validated as of the signing
	// time recorded in that evidence rather than the current time.
	if envelopeHasCertificate(envelope) {
		if bundlePath := resolveSigstoreBundlePath(policy, path); bundlePath != "" && bundleAnchorsSigningTime(bundlePath) {
			policy.SigstoreBundlePath = bundlePath
			return verifySigstoreBundleEvidence(envelope, cfg, policy)
		}
	}

	verifier, err := verifierForEnvelope(ctx, cfg, envelope)
	if err != nil {
		return err
	}

	if err := VerifyEnvelope(ctx, envelope, verifier); err != nil {
		return fmt.Errorf("verify attestation envelope %q: %w", path, err)
	}

	return nil
}

// resolveSigstoreBundlePath returns an explicitly configured bundle path, or the
// conventional "<attestation>.sigstore.json" sidecar written alongside signed
// attestations when it exists on disk.
func resolveSigstoreBundlePath(policy VerificationPolicy, attestationPath string) string {
	if trimmed := strings.TrimSpace(policy.SigstoreBundlePath); trimmed != "" {
		return trimmed
	}

	candidate := defaultSigstoreBundlePath(attestationPath)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}

	return ""
}

func defaultSigstoreBundlePath(attestationPath string) string {
	if strings.HasSuffix(attestationPath, ".json") {
		return strings.TrimSuffix(attestationPath, ".json") + ".sigstore.json"
	}

	return attestationPath + ".sigstore.json"
}

// bundleAnchorsSigningTime reports whether the bundle carries a trusted time
// source (a transparency-log entry or an RFC3161 timestamp) that can validate
// the signing certificate as of signing time.
var bundleAnchorsSigningTime = func(bundlePath string) bool {
	bundle, err := loadSigstoreBundle(bundlePath)
	if err != nil {
		return false
	}

	if entries, err := bundle.TlogEntries(); err == nil && len(entries) > 0 {
		return true
	}

	if timestamps, err := bundle.Timestamps(); err == nil && len(timestamps) > 0 {
		return true
	}

	return false
}

func verifierForEnvelope(ctx context.Context, cfg BackendConfig, envelope Envelope) (Verifier, error) {
	mode := normalizeVerificationMode(cfg, envelope)

	if cfg.VerifierRef != "" {
		if mode == SigningModeKeyless || envelopeHasCertificate(envelope) {
			return nil, fmt.Errorf("verifier overrides are not supported for keyless attestations; verify with keyless_identity and keyless_oidc_issuer instead")
		}
		return LoadPEMVerifier(cfg.VerifierRef)
	}

	switch mode {
	case SigningModeKey:
		if cfg.SignerRef == "" {
			return nil, fmt.Errorf("attestation verification for signing_mode %q requires verifier or key", SigningModeKey)
		}
		return LoadPEMVerifier(cfg.SignerRef)
	case SigningModeKMS:
		if cfg.SignerRef == "" {
			return nil, fmt.Errorf("attestation verification for signing_mode %q requires key or verifier", SigningModeKMS)
		}
		signer, err := NewSigner(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return signer.Verifier(ctx, cfg)
	case SigningModeKeyless:
		return newKeylessVerifierForEnvelope(cfg, envelope)
	default:
		return nil, fmt.Errorf("unable to determine attestation signing mode; set signing_mode or verifier explicitly")
	}
}

func normalizeVerificationMode(cfg BackendConfig, envelope Envelope) string {
	if cfg.Mode != "" {
		return cfg.Mode
	}

	if envelopeHasCertificate(envelope) {
		return SigningModeKeyless
	}

	if isRecognizedKMSReference(cfg.SignerRef) {
		return SigningModeKMS
	}

	if cfg.SignerRef != "" || cfg.VerifierRef != "" {
		return SigningModeKey
	}

	return ""
}

func envelopeHasCertificate(envelope Envelope) bool {
	for _, signature := range envelope.Signatures {
		if strings.TrimSpace(signature.Cert) != "" {
			return true
		}
	}

	return false
}

func isRecognizedKMSReference(value string) bool {
	for _, prefix := range []string{"awskms://", "gcpkms://", "azurekms://", "hashivault://"} {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}

func verifyPolicy(payload []byte, policy VerificationPolicy) (*internalprovenance.Statement, error) {
	var statement internalprovenance.Statement
	if err := json.Unmarshal(payload, &statement); err != nil {
		return nil, fmt.Errorf("decode attestation statement: %w", err)
	}

	if statement.Type != internalprovenance.StatementType {
		return nil, fmt.Errorf("unexpected attestation statement type %q", statement.Type)
	}

	if policy.PredicateType != "" && statement.PredicateType != policy.PredicateType {
		return nil, fmt.Errorf("attestation predicate type %q does not match expected %q", statement.PredicateType, policy.PredicateType)
	}

	if policy.ArtifactPath != "" {
		if err := verifyArtifactSubject(statement.Subject, policy.ArtifactPath); err != nil {
			return nil, err
		}
	}

	if policy.BuilderID != "" || policy.SourceURI != "" {
		if statement.PredicateType != internalprovenance.SLSAProvenanceV1PredicateType {
			return nil, fmt.Errorf("builder and source policy checks require predicate type %q, got %q", internalprovenance.SLSAProvenanceV1PredicateType, statement.PredicateType)
		}

		var typedStatement struct {
			Type          string                                     `json:"_type"`
			Subject       []internalprovenance.Subject               `json:"subject"`
			PredicateType string                                     `json:"predicateType"`
			Predicate     internalprovenance.SLSAProvenancePredicate `json:"predicate"`
		}
		if err := json.Unmarshal(payload, &typedStatement); err != nil {
			return nil, fmt.Errorf("decode SLSA predicate for policy verification: %w", err)
		}

		if policy.BuilderID != "" && typedStatement.Predicate.RunDetails.Builder.ID != policy.BuilderID {
			return nil, fmt.Errorf("attestation builder id %q does not match expected %q", typedStatement.Predicate.RunDetails.Builder.ID, policy.BuilderID)
		}

		if policy.SourceURI != "" {
			matched := false
			for _, dependency := range typedStatement.Predicate.BuildDefinition.ResolvedDependencies {
				if dependency.URI == policy.SourceURI {
					matched = true
					break
				}
			}
			if !matched {
				return nil, fmt.Errorf("attestation does not contain expected source URI %q", policy.SourceURI)
			}
		}
	}

	return &statement, nil
}

func requiresSigstoreBundle(policy VerificationPolicy) bool {
	return policy.RequireTransparencyLog || policy.RequireObserverTimestamp
}

func verifySigstoreBundleEvidenceImpl(envelope Envelope, cfg BackendConfig, policy VerificationPolicy) error {
	if strings.TrimSpace(policy.SigstoreBundlePath) == "" {
		return fmt.Errorf("bundle-based Rekor or timestamp verification requires -bundle")
	}

	if normalizeVerificationMode(cfg, envelope) != SigningModeKeyless && !envelopeHasCertificate(envelope) {
		return fmt.Errorf("bundle-based Rekor or timestamp verification currently requires a keyless attestation")
	}

	if strings.TrimSpace(cfg.KeylessIdentity) == "" || strings.TrimSpace(cfg.KeylessOIDCIssuer) == "" {
		return fmt.Errorf("bundle-based Rekor or timestamp verification requires keyless_identity and keyless_oidc_issuer")
	}

	trustedMaterial, err := loadKeylessTrustedMaterial(cfg)
	if err != nil {
		return fmt.Errorf("load keyless trusted root: %w", err)
	}

	bundle, err := loadSigstoreBundle(policy.SigstoreBundlePath)
	if err != nil {
		return fmt.Errorf("load Sigstore bundle %q: %w", policy.SigstoreBundlePath, err)
	}

	if err := ensureBundleMatchesEnvelope(bundle, envelope); err != nil {
		return err
	}

	verifierOptions := []sigstoreverify.VerifierOption{}
	if policy.RequireTransparencyLog {
		verifierOptions = append(verifierOptions, sigstoreverify.WithTransparencyLog(1))
	}
	if policy.RequireObserverTimestamp {
		verifierOptions = append(verifierOptions, sigstoreverify.WithObserverTimestamps(1))
	}
	if len(verifierOptions) == 0 {
		// A trusted time source is required to validate the short-lived Fulcio
		// certificate as of signing time; default to observer timestamps when the
		// caller has not explicitly required Rekor or timestamp evidence.
		verifierOptions = append(verifierOptions, sigstoreverify.WithObserverTimestamps(1))
	}

	verifier, err := newSigstoreBundleVerifier(trustedMaterial, verifierOptions...)
	if err != nil {
		return fmt.Errorf("create Sigstore bundle verifier: %w", err)
	}

	artifactPolicy := sigstoreverify.WithoutArtifactUnsafe()
	if policy.ArtifactPath != "" {
		artifact, err := os.Open(policy.ArtifactPath)
		if err != nil {
			return fmt.Errorf("open artifact %q for bundle verification: %w", policy.ArtifactPath, err)
		}
		defer func() { _ = artifact.Close() }()
		artifactPolicy = sigstoreverify.WithArtifact(artifact)
	}

	identity, err := sigstoreverify.NewShortCertificateIdentity(cfg.KeylessOIDCIssuer, "", cfg.KeylessIdentity, "")
	if err != nil {
		return fmt.Errorf("build keyless identity policy: %w", err)
	}

	policyBuilder := sigstoreverify.NewPolicy(artifactPolicy, sigstoreverify.WithCertificateIdentity(identity))
	if _, err := verifier.Verify(bundle, policyBuilder); err != nil {
		return fmt.Errorf("verify Sigstore bundle %q: %w", policy.SigstoreBundlePath, err)
	}

	return nil
}

func ensureBundleMatchesEnvelope(bundle *sigstorebundle.Bundle, envelope Envelope) error {
	bundleEnvelope, err := bundle.Envelope()
	if err != nil {
		return fmt.Errorf("extract DSSE envelope from Sigstore bundle: %w", err)
	}

	rawEnvelope := bundleEnvelope.RawEnvelope()
	if rawEnvelope == nil {
		return fmt.Errorf("sigstore bundle does not contain a DSSE envelope")
	}

	if rawEnvelope.PayloadType != envelope.PayloadType || rawEnvelope.Payload != envelope.Payload {
		return fmt.Errorf("sigstore bundle payload does not match attestation")
	}

	if len(envelope.Signatures) == 0 {
		return fmt.Errorf("attestation envelope has no signatures")
	}

	bundleSignature := bundleEnvelope.Signature()
	for i, envelopeSignature := range envelope.Signatures {
		signature, err := DecodeEnvelopeSignature(envelopeSignature)
		if err != nil {
			return fmt.Errorf("decode attestation envelope signature %d: %w", i, err)
		}

		if bytes.Equal(bundleSignature, signature) {
			return nil
		}
	}

	return fmt.Errorf("sigstore bundle signature does not match any attestation signature")
}

func verifyArtifactSubject(subjects []internalprovenance.Subject, artifactPath string) error {
	digest, err := sha256File(artifactPath)
	if err != nil {
		return fmt.Errorf("hash artifact %q: %w", artifactPath, err)
	}

	artifactName := filepath.Base(artifactPath)
	for _, subject := range subjects {
		if subject.Name == artifactName && strings.EqualFold(subject.Digest["sha256"], digest) {
			return nil
		}
	}

	return fmt.Errorf("attestation subject does not match artifact %q", artifactPath)
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
