// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"bytes"
	"context"
	"crypto"
	"errors"
	"fmt"
	"strings"

	sigstoresignature "github.com/sigstore/sigstore/pkg/signature"
	sigstorekms "github.com/sigstore/sigstore/pkg/signature/kms"
)

var newKMSSignerVerifier = func(ctx context.Context, keyResourceID string) (sigstorekms.SignerVerifier, error) {
	return sigstorekms.Get(ctx, keyResourceID, crypto.SHA256)
}

func init() {
	RegisterSigner(SigningModeKMS, newKMSSigner)
}

type kmsSigner struct {
	signerVerifier sigstorekms.SignerVerifier
	verifier       Verifier
	keyID          string
}

func newKMSSigner(ctx context.Context, cfg BackendConfig) (Signer, error) {
	if cfg.SignerRef == "" {
		return nil, fmt.Errorf("signing_mode %q requires signer or key", SigningModeKMS)
	}

	signerVerifier, err := newKMSSignerVerifier(ctx, cfg.SignerRef)
	if err != nil {
		var notFound *sigstorekms.ProviderNotFoundError
		if errors.As(err, &notFound) {
			return nil, fmt.Errorf("initialize KMS signer %q: %w%s", cfg.SignerRef, err, kmsProviderBuildHint(cfg.SignerRef))
		}
		return nil, fmt.Errorf("initialize KMS signer %q: %w", cfg.SignerRef, err)
	}

	publicKey, err := signerVerifier.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("load KMS public key %q: %w", cfg.SignerRef, err)
	}

	verifier, err := newSigstoreVerifierFromPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("create KMS verifier %q: %w", cfg.SignerRef, err)
	}

	return &kmsSigner{
		signerVerifier: signerVerifier,
		verifier:       verifier,
		keyID:          verifier.KeyID(),
	}, nil
}

func (s *kmsSigner) Sign(_ context.Context, payloadType string, payload []byte) (Signature, error) {
	encoded := PreAuthEncode(payloadType, payload)
	signature, err := s.signerVerifier.SignMessage(bytes.NewReader(encoded))
	if err != nil {
		return Signature{}, fmt.Errorf("sign payload with KMS: %w", err)
	}

	return Signature{
		KeyID: s.keyID,
		Sig:   signature,
	}, nil
}

func (s *kmsSigner) Verifier(context.Context, BackendConfig) (Verifier, error) {
	return s.verifier, nil
}

// kmsProviderBuildTags maps a KMS/Vault URI scheme to the build tag that
// compiles its provider into the binary. Providers are included by default;
// builds using the "kms_cherrypick" tag opt in to individual providers.
var kmsProviderBuildTags = map[string]string{
	"awskms":     "kms_aws",
	"gcpkms":     "kms_gcp",
	"azurekms":   "kms_azure",
	"hashivault": "kms_hashivault",
}

// kmsProviderBuildHint returns guidance when a recognized KMS provider scheme is
// requested but no provider is registered, which happens when the binary was
// built with "kms_cherrypick" and did not opt that provider in.
func kmsProviderBuildHint(ref string) string {
	scheme := ref
	if idx := strings.Index(ref, "://"); idx >= 0 {
		scheme = ref[:idx]
	}

	tag, ok := kmsProviderBuildTags[scheme]
	if !ok {
		return ""
	}

	return fmt.Sprintf("; the %s KMS provider is not compiled into this build (rebuild without \"kms_cherrypick\", or with -tags 'kms_cherrypick %s')", scheme, tag)
}

func newSigstoreVerifierFromPublicKey(publicKey crypto.PublicKey) (Verifier, error) {
	verifier, err := sigstoresignature.LoadDefaultVerifier(publicKey)
	if err != nil {
		return nil, fmt.Errorf("load sigstore verifier: %w", err)
	}

	publicKeyPEM, err := marshalPublicKeyPEM(publicKey)
	if err != nil {
		return nil, err
	}

	return &sigstoreVerifier{
		verifier: verifier,
		keyID:    sha256Hex(publicKeyPEM),
	}, nil
}

type sigstoreVerifier struct {
	verifier sigstoresignature.Verifier
	keyID    string
}

func (v *sigstoreVerifier) Verify(_ context.Context, payloadType string, payload, signature []byte) error {
	encoded := PreAuthEncode(payloadType, payload)
	if err := v.verifier.VerifySignature(bytes.NewReader(signature), bytes.NewReader(encoded)); err != nil {
		return err
	}

	return nil
}

func (v *sigstoreVerifier) KeyID() string {
	return v.keyID
}
