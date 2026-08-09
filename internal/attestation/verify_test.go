// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"encoding/base64"
	"strings"
	"testing"

	protobundle "github.com/sigstore/protobuf-specs/gen/pb-go/bundle/v1"
	protodsse "github.com/sigstore/protobuf-specs/gen/pb-go/dsse"
	sigstorebundle "github.com/sigstore/sigstore-go/pkg/bundle"
)

// newDSSEBundle builds a minimal Sigstore bundle wrapping a DSSE envelope with
// the provided raw payload and signature. It bypasses bundle.NewBundle because
// that constructor requires full verification material; ensureBundleMatchesEnvelope
// only inspects the DSSE content.
func newDSSEBundle(payload, signature []byte) *sigstorebundle.Bundle {
	return &sigstorebundle.Bundle{
		Bundle: &protobundle.Bundle{
			Content: &protobundle.Bundle_DsseEnvelope{
				DsseEnvelope: &protodsse.Envelope{
					Payload:     payload,
					PayloadType: InTotoPayloadType,
					Signatures:  []*protodsse.Signature{{Sig: signature}},
				},
			},
		},
	}
}

func TestEnsureBundleMatchesEnvelopeMatchesNonFirstSignature(t *testing.T) {
	payload := []byte(`{"hello":"world"}`)
	bundleSignature := []byte("the-real-signature")

	bundle := newDSSEBundle(payload, bundleSignature)

	// The bundle signature is the second envelope signature, not the first.
	envelope := Envelope{
		PayloadType: InTotoPayloadType,
		Payload:     base64.StdEncoding.EncodeToString(payload),
		Signatures: []EnvelopeSignature{
			{Sig: base64.StdEncoding.EncodeToString([]byte("a-different-signature"))},
			{Sig: base64.StdEncoding.EncodeToString(bundleSignature)},
		},
	}

	if err := ensureBundleMatchesEnvelope(bundle, envelope); err != nil {
		t.Fatalf("expected bundle to match a later envelope signature, got error: %v", err)
	}
}

func TestEnsureBundleMatchesEnvelopeRejectsWhenNoSignatureMatches(t *testing.T) {
	payload := []byte(`{"hello":"world"}`)
	bundle := newDSSEBundle(payload, []byte("the-real-signature"))

	envelope := Envelope{
		PayloadType: InTotoPayloadType,
		Payload:     base64.StdEncoding.EncodeToString(payload),
		Signatures: []EnvelopeSignature{
			{Sig: base64.StdEncoding.EncodeToString([]byte("a-different-signature"))},
			{Sig: base64.StdEncoding.EncodeToString([]byte("another-mismatch"))},
		},
	}

	err := ensureBundleMatchesEnvelope(bundle, envelope)
	if err == nil {
		t.Fatalf("expected mismatch error when no envelope signature matches the bundle")
	}
	if !strings.Contains(err.Error(), "does not match any attestation signature") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureBundleMatchesEnvelopeRejectsPayloadMismatch(t *testing.T) {
	bundle := newDSSEBundle([]byte(`{"hello":"world"}`), []byte("sig"))

	envelope := Envelope{
		PayloadType: InTotoPayloadType,
		Payload:     base64.StdEncoding.EncodeToString([]byte(`{"hello":"tampered"}`)),
		Signatures: []EnvelopeSignature{
			{Sig: base64.StdEncoding.EncodeToString([]byte("sig"))},
		},
	}

	err := ensureBundleMatchesEnvelope(bundle, envelope)
	if err == nil {
		t.Fatalf("expected payload mismatch to be rejected")
	}
	if !strings.Contains(err.Error(), "payload does not match") {
		t.Fatalf("unexpected error: %v", err)
	}
}
