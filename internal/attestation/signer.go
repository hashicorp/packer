// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"context"
	"fmt"
)

const (
	SigningModeNone    = "none"
	SigningModeKey     = "key"
	SigningModeKMS     = "kms"
	SigningModeKeyless = "keyless"
)

type Signature struct {
	KeyID   string
	Sig     []byte
	CertPEM []byte
}

type Signer interface {
	Sign(ctx context.Context, payloadType string, payload []byte) (Signature, error)
	Verifier(ctx context.Context, cfg BackendConfig) (Verifier, error)
}

type Verifier interface {
	Verify(ctx context.Context, payloadType string, payload, signature []byte) error
	KeyID() string
}

type BackendConfig struct {
	Mode              string
	SignerRef         string
	VerifierRef       string
	Env               map[string]string
	FulcioURL         string
	RekorURL          string
	UploadTlog        bool
	TrustedRootPath   string
	KeylessIdentity   string
	KeylessOIDCIssuer string
}

type signerFactory func(context.Context, BackendConfig) (Signer, error)

var signerFactories = map[string]signerFactory{}

func RegisterSigner(mode string, factory signerFactory) {
	signerFactories[mode] = factory
}

func NewSigner(ctx context.Context, cfg BackendConfig) (Signer, error) {
	factory, ok := signerFactories[cfg.Mode]
	if !ok {
		return nil, fmt.Errorf("signing_mode %q is not implemented", cfg.Mode)
	}

	return factory(ctx, cfg)
}

func NewVerifier(ctx context.Context, cfg BackendConfig, signer Signer) (Verifier, error) {
	if cfg.VerifierRef != "" {
		return LoadPEMVerifier(cfg.VerifierRef)
	}

	return signer.Verifier(ctx, cfg)
}

func VerifyEnvelope(ctx context.Context, envelope Envelope, verifier Verifier) error {
	if len(envelope.Signatures) == 0 {
		return fmt.Errorf("envelope has no signatures")
	}

	payload, err := DecodeEnvelopePayload(envelope)
	if err != nil {
		return err
	}

	for _, signature := range envelope.Signatures {
		decodedSignature, decodeErr := DecodeEnvelopeSignature(signature)
		if decodeErr != nil {
			return decodeErr
		}

		if verifyErr := verifier.Verify(ctx, envelope.PayloadType, payload, decodedSignature); verifyErr == nil {
			return nil
		}
	}

	return fmt.Errorf("signature verification failed")
}
