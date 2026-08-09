// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"context"
	"fmt"
)

type bundleSigner interface {
	SignBundle(ctx context.Context, payloadType string, payload []byte, cfg BackendConfig) (Envelope, []byte, error)
}

func BuildBundleForSigner(ctx context.Context, signer Signer, cfg BackendConfig, payloadType string, payload []byte) (Envelope, []byte, error) {
	bundler, ok := signer.(bundleSigner)
	if !ok {
		return Envelope{}, nil, fmt.Errorf("signing_mode %q does not support Sigstore bundle emission", cfg.Mode)
	}

	return bundler.SignBundle(ctx, payloadType, payload, cfg)
}
