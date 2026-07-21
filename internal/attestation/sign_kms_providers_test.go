// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

//go:build !kms_cherrypick

package attestation

import (
	"strings"
	"testing"

	sigstorekms "github.com/sigstore/sigstore/pkg/signature/kms"
)

func TestDefaultBuildRegistersAllKMSProviders(t *testing.T) {
	registered := map[string]bool{}
	for _, scheme := range sigstorekms.SupportedProviders() {
		registered[scheme] = true
	}

	for _, scheme := range []string{"awskms://", "azurekms://", "gcpkms://", "hashivault://"} {
		if !registered[scheme] {
			t.Errorf("default build is missing KMS provider %q (registered: %v)", scheme, sigstorekms.SupportedProviders())
		}
	}
}

func TestKMSProviderBuildHint(t *testing.T) {
	hint := kmsProviderBuildHint("awskms://alias/example")
	if !strings.Contains(hint, "kms_aws") || !strings.Contains(hint, "awskms") {
		t.Fatalf("unexpected hint for awskms reference: %q", hint)
	}

	if got := kmsProviderBuildHint("file:///tmp/key.pem"); got != "" {
		t.Fatalf("expected no build hint for non-KMS reference, got %q", got)
	}
}
