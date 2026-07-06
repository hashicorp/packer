// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package attestation

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	sigstorebundle "github.com/sigstore/sigstore-go/pkg/bundle"
	fulciocertificate "github.com/sigstore/sigstore-go/pkg/fulcio/certificate"
	sigstoreroot "github.com/sigstore/sigstore-go/pkg/root"
	sigstoregosign "github.com/sigstore/sigstore-go/pkg/sign"
	sigstoreverify "github.com/sigstore/sigstore-go/pkg/verify"
)

const defaultFulcioURL = "https://fulcio.sigstore.dev"
const defaultRekorURL = "https://rekor.sigstore.dev"

var newKeylessEphemeralKeypair = func() (sigstoregosign.Keypair, error) {
	return sigstoregosign.NewEphemeralKeypair(nil)
}

var newKeylessFulcio = func(baseURL string) sigstoregosign.CertificateProvider {
	return sigstoregosign.NewFulcio(&sigstoregosign.FulcioOptions{BaseURL: baseURL})
}

var newKeylessBundle = sigstoregosign.Bundle

var newKeylessRekor = func(baseURL string) sigstoregosign.Transparency {
	return sigstoregosign.NewRekor(&sigstoregosign.RekorOptions{BaseURL: baseURL})
}

var loadKeylessTrustedMaterial = func(cfg BackendConfig) (sigstoreroot.TrustedMaterial, error) {
	trustedRootPath := strings.TrimSpace(cfg.TrustedRootPath)
	if trustedRootPath == "" {
		return sigstoreroot.FetchTrustedRoot()
	}

	return sigstoreroot.NewTrustedRootFromPath(trustedRootPath)
}

var verifyKeylessCertificate = func(certificate *x509.Certificate, trustedMaterial sigstoreroot.TrustedMaterial, expectedIdentity, expectedOIDCIssuer string) error {
	if _, err := sigstoreverify.VerifyLeafCertificate(time.Now().UTC(), certificate, trustedMaterial); err != nil {
		return fmt.Errorf("verify Fulcio certificate chain: %w", err)
	}

	summary, err := fulciocertificate.SummarizeCertificate(certificate)
	if err != nil {
		return fmt.Errorf("summarize Fulcio certificate: %w", err)
	}

	identity, err := sigstoreverify.NewShortCertificateIdentity(expectedOIDCIssuer, "", expectedIdentity, "")
	if err != nil {
		return fmt.Errorf("build keyless identity policy: %w", err)
	}
	if err := identity.Verify(summary); err != nil {
		return fmt.Errorf("verify keyless certificate identity: %w", err)
	}

	return nil
}

func init() {
	RegisterSigner(SigningModeKeyless, newKeylessSigner)
}

type keylessSigner struct {
	keypair  sigstoregosign.Keypair
	certPEM  []byte
	cert     *x509.Certificate
	verifier Verifier
	keyID    string
}

func newKeylessSigner(ctx context.Context, cfg BackendConfig) (Signer, error) {
	fulcioURL := strings.TrimSpace(cfg.FulcioURL)
	if fulcioURL == "" {
		fulcioURL = defaultFulcioURL
	}

	idToken, err := resolveAmbientIDToken(cfg.Env)
	if err != nil {
		return nil, err
	}

	keypair, err := newKeylessEphemeralKeypair()
	if err != nil {
		return nil, fmt.Errorf("generate ephemeral keypair: %w", err)
	}

	fulcio := newKeylessFulcio(fulcioURL)
	certDER, err := fulcio.GetCertificate(ctx, keypair, &sigstoregosign.CertificateProviderOptions{IDToken: idToken})
	if err != nil {
		return nil, fmt.Errorf("request Fulcio certificate: %w", err)
	}

	certificate, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parse Fulcio certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	verifier, err := newSigstoreVerifierFromPublicKey(certificate.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("create keyless verifier: %w", err)
	}

	return &keylessSigner{
		keypair:  keypair,
		certPEM:  certPEM,
		cert:     certificate,
		verifier: verifier,
		keyID:    hex.EncodeToString(keypair.GetHint()),
	}, nil
}

func (s *keylessSigner) Sign(ctx context.Context, payloadType string, payload []byte) (Signature, error) {
	signature, _, err := s.keypair.SignData(ctx, PreAuthEncode(payloadType, payload))
	if err != nil {
		return Signature{}, fmt.Errorf("sign payload with keyless signer: %w", err)
	}

	return Signature{
		KeyID:   s.keyID,
		Sig:     signature,
		CertPEM: append([]byte(nil), s.certPEM...),
	}, nil
}

func (s *keylessSigner) SignBundle(ctx context.Context, payloadType string, payload []byte, cfg BackendConfig) (Envelope, []byte, error) {
	content := &sigstoregosign.DSSEData{Data: payload, PayloadType: payloadType}
	options := sigstoregosign.BundleOptions{
		CertificateProvider: staticCertificateProvider{certDER: append([]byte(nil), s.cert.Raw...)},
		Context:             ctx,
	}

	if cfg.UploadTlog {
		rekorURL := strings.TrimSpace(cfg.RekorURL)
		if rekorURL == "" {
			rekorURL = defaultRekorURL
		}

		trustedMaterial, err := loadKeylessTrustedMaterial(cfg)
		if err != nil {
			return Envelope{}, nil, fmt.Errorf("load keyless trusted root: %w", err)
		}

		options.TransparencyLogs = []sigstoregosign.Transparency{newKeylessRekor(rekorURL)}
		options.TrustedRoot = trustedMaterial
	}

	protobufBundle, err := newKeylessBundle(content, s.keypair, options)
	if err != nil {
		return Envelope{}, nil, fmt.Errorf("build Sigstore bundle: %w", err)
	}

	bundleWrapper, err := sigstorebundle.NewBundle(protobufBundle)
	if err != nil {
		return Envelope{}, nil, fmt.Errorf("decode Sigstore bundle: %w", err)
	}

	bundleEnvelope, err := bundleWrapper.Envelope()
	if err != nil {
		return Envelope{}, nil, fmt.Errorf("extract envelope from Sigstore bundle: %w", err)
	}

	rawEnvelope := bundleEnvelope.RawEnvelope()
	if rawEnvelope == nil {
		return Envelope{}, nil, fmt.Errorf("sigstore bundle does not contain a DSSE envelope")
	}

	bundleJSON, err := bundleWrapper.MarshalJSON()
	if err != nil {
		return Envelope{}, nil, fmt.Errorf("marshal Sigstore bundle: %w", err)
	}

	envelope := Envelope{
		PayloadType: rawEnvelope.PayloadType,
		Payload:     rawEnvelope.Payload,
		Signatures: []EnvelopeSignature{{
			KeyID: s.keyID,
			Sig:   base64.StdEncoding.EncodeToString(bundleEnvelope.Signature()),
			Cert:  string(s.certPEM),
		}},
	}

	return envelope, bundleJSON, nil
}

func (s *keylessSigner) Verifier(ctx context.Context, cfg BackendConfig) (Verifier, error) {
	return newKeylessVerifier(cfg, s.cert)
}

func newKeylessVerifierForEnvelope(cfg BackendConfig, envelope Envelope) (Verifier, error) {
	certificate, err := certificateFromEnvelope(envelope)
	if err != nil {
		return nil, err
	}

	return newKeylessVerifier(cfg, certificate)
}

func newKeylessVerifier(cfg BackendConfig, certificate *x509.Certificate) (Verifier, error) {
	if strings.TrimSpace(cfg.KeylessIdentity) == "" || strings.TrimSpace(cfg.KeylessOIDCIssuer) == "" {
		return nil, fmt.Errorf("signing_mode %q requires keyless_identity and keyless_oidc_issuer unless verifier is explicitly configured", SigningModeKeyless)
	}

	trustedMaterial, err := loadKeylessTrustedMaterial(cfg)
	if err != nil {
		return nil, fmt.Errorf("load keyless trusted root: %w", err)
	}

	signatureVerifier, err := newSigstoreVerifierFromPublicKey(certificate.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("create keyless verifier: %w", err)
	}

	return &keylessVerifier{
		certificate:       certificate,
		signatureVerifier: signatureVerifier,
		trustedMaterial:   trustedMaterial,
		identity:          cfg.KeylessIdentity,
		oidcIssuer:        cfg.KeylessOIDCIssuer,
	}, nil
}

type keylessVerifier struct {
	certificate       *x509.Certificate
	signatureVerifier Verifier
	trustedMaterial   sigstoreroot.TrustedMaterial
	identity          string
	oidcIssuer        string
}

func (v *keylessVerifier) Verify(ctx context.Context, payloadType string, payload, signature []byte) error {
	if err := verifyKeylessCertificate(v.certificate, v.trustedMaterial, v.identity, v.oidcIssuer); err != nil {
		return err
	}

	return v.signatureVerifier.Verify(ctx, payloadType, payload, signature)
}

func (v *keylessVerifier) KeyID() string {
	return v.signatureVerifier.KeyID()
}

func certificateFromEnvelope(envelope Envelope) (*x509.Certificate, error) {
	for _, signature := range envelope.Signatures {
		if strings.TrimSpace(signature.Cert) == "" {
			continue
		}

		block, _ := pem.Decode([]byte(signature.Cert))
		if block == nil {
			return nil, fmt.Errorf("decode keyless certificate: no PEM block found")
		}

		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse keyless certificate: %w", err)
		}

		return certificate, nil
	}

	return nil, fmt.Errorf("keyless attestation does not contain a signing certificate")
}

type staticCertificateProvider struct {
	certDER []byte
}

func (p staticCertificateProvider) GetCertificate(context.Context, sigstoregosign.Keypair, *sigstoregosign.CertificateProviderOptions) ([]byte, error) {
	if len(p.certDER) == 0 {
		return nil, fmt.Errorf("static certificate provider is missing a certificate")
	}

	return append([]byte(nil), p.certDER...), nil
}

func resolveAmbientIDToken(env map[string]string) (string, error) {
	if token := strings.TrimSpace(env["SIGSTORE_ID_TOKEN"]); token != "" {
		return token, nil
	}
	if token := strings.TrimSpace(env["CI_JOB_JWT_V2"]); token != "" {
		return token, nil
	}
	if token := strings.TrimSpace(env["CI_JOB_JWT"]); token != "" {
		return token, nil
	}
	if token, err := resolveGitHubActionsIDToken(env); err == nil && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("signing_mode %q requires an ambient OIDC token; set SIGSTORE_ID_TOKEN, use a CI-provided OIDC token, or switch to signing_mode=\"none\" or a key-backed mode", SigningModeKeyless)
}

func resolveGitHubActionsIDToken(env map[string]string) (string, error) {
	requestURL := strings.TrimSpace(env["ACTIONS_ID_TOKEN_REQUEST_URL"])
	requestToken := strings.TrimSpace(env["ACTIONS_ID_TOKEN_REQUEST_TOKEN"])
	if requestURL == "" || requestToken == "" {
		return "", nil
	}

	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("parse GitHub OIDC request URL: %w", err)
	}
	query := parsedURL.Query()
	if query.Get("audience") == "" {
		query.Set("audience", "sigstore")
		parsedURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create GitHub OIDC request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request GitHub OIDC token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("request GitHub OIDC token: unexpected status %s", resp.Status)
	}

	var payload struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode GitHub OIDC token response: %w", err)
	}
	return strings.TrimSpace(payload.Value), nil
}
