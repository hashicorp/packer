// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const InTotoPayloadType = "application/vnd.in-toto+json"

type Envelope struct {
	PayloadType string              `json:"payloadType"`
	Payload     string              `json:"payload"`
	Signatures  []EnvelopeSignature `json:"signatures"`
}

type EnvelopeSignature struct {
	KeyID string `json:"keyid,omitempty"`
	Sig   string `json:"sig"`
	Cert  string `json:"cert,omitempty"`
}

func MarshalPayload(value any) ([]byte, error) {
	return json.Marshal(value)
}

func PreAuthEncode(payloadType string, payload []byte) []byte {
	return []byte(fmt.Sprintf("DSSEv1 %d %s %d %s", len(payloadType), payloadType, len(payload), payload))
}

func NewEnvelope(payloadType string, payload []byte, signature Signature) Envelope {
	envelope := Envelope{
		PayloadType: payloadType,
		Payload:     base64.StdEncoding.EncodeToString(payload),
		Signatures: []EnvelopeSignature{{
			KeyID: signature.KeyID,
			Sig:   base64.StdEncoding.EncodeToString(signature.Sig),
		}},
	}

	if len(signature.CertPEM) > 0 {
		envelope.Signatures[0].Cert = string(signature.CertPEM)
	}

	return envelope
}

func DecodeEnvelopePayload(envelope Envelope) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		return nil, fmt.Errorf("decode envelope payload: %w", err)
	}

	return decoded, nil
}

func DecodeEnvelopeSignature(signature EnvelopeSignature) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(signature.Sig)
	if err != nil {
		return nil, fmt.Errorf("decode envelope signature: %w", err)
	}

	return decoded, nil
}
