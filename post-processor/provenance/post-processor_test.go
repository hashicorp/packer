// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	filebuilder "github.com/hashicorp/packer/builder/file"
	internalattestation "github.com/hashicorp/packer/internal/attestation"
	internalprovenance "github.com/hashicorp/packer/internal/provenance"
	internalsbom "github.com/hashicorp/packer/internal/sbom"
)

func TestPostProcessorWritesUnsignedStatementAndPreservesArtifact(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	outputDir := t.TempDir()
	config := mustTemplateJSON(t, map[string]interface{}{
		"post-processors": []map[string]string{{
			"type":       "provenance",
			"output_dir": outputDir,
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var postProcessor PostProcessor
	if err := postProcessor.Configure(tpl.PostProcessors[0][0].Config); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}

	returnedArtifact, keep, mustKeep, err := postProcessor.PostProcess(context.Background(), packersdk.TestUi(t), artifact)
	if err != nil {
		t.Fatalf("post-process artifact: %v", err)
	}

	if returnedArtifact != artifact {
		t.Fatalf("expected original artifact to be preserved")
	}
	if !keep || !mustKeep {
		t.Fatalf("expected keep and mustKeep to be true")
	}

	statementPath := filepath.Join(outputDir, "package.txt.provenance.json")
	contents, err := os.ReadFile(statementPath)
	if err != nil {
		t.Fatalf("read provenance statement: %v", err)
	}

	var statement internalprovenance.Statement
	if err := json.Unmarshal(contents, &statement); err != nil {
		t.Fatalf("unmarshal statement: %v", err)
	}

	if got, want := statement.Type, internalprovenance.StatementType; got != want {
		t.Fatalf("unexpected statement type %q, want %q", got, want)
	}
	if got, want := statement.PredicateType, internalprovenance.SLSAProvenanceV1PredicateType; got != want {
		t.Fatalf("unexpected predicate type %q, want %q", got, want)
	}
	if got, want := len(statement.Subject), 1; got != want {
		t.Fatalf("unexpected subject count %d, want %d", got, want)
	}
}

func TestPostProcessorEnrichesPredicateFromConfigAndCIEnv(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	outputDir := t.TempDir()
	config := mustTemplateJSON(t, map[string]interface{}{
		"post-processors": []map[string]interface{}{{
			"type":           "provenance",
			"output_dir":     outputDir,
			"template":       "ubuntu.pkr.hcl",
			"only_builds":    []string{"qemu.ubuntu"},
			"user_variables": map[string]string{"role": "web"},
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var postProcessor PostProcessor
	if err := postProcessor.Configure(tpl.PostProcessors[0][0].Config); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}

	times := []time.Time{
		time.Date(2026, time.July, 4, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.July, 4, 10, 12, 0, 0, time.UTC),
	}
	timeIndex := 0
	postProcessor.now = func() time.Time {
		current := times[timeIndex]
		if timeIndex < len(times)-1 {
			timeIndex++
		}
		return current
	}
	postProcessor.env = map[string]string{
		"GITHUB_REPOSITORY":   "acme/images",
		"GITHUB_SHA":          "deadbeef",
		"GITHUB_REF":          "refs/heads/main",
		"GITHUB_WORKFLOW_REF": "acme/images/.github/workflows/build.yml@refs/heads/main",
		"GITHUB_RUN_ID":       "run-42",
		"PKR_VAR_region":      "us-east-1",
	}
	postProcessor.workingDir = "/workspace/packer"

	_, _, _, err = postProcessor.PostProcess(context.Background(), packersdk.TestUi(t), artifact)
	if err != nil {
		t.Fatalf("post-process artifact: %v", err)
	}

	statementPath := filepath.Join(outputDir, "package.txt.provenance.json")
	contents, err := os.ReadFile(statementPath)
	if err != nil {
		t.Fatalf("read provenance statement: %v", err)
	}

	var statement struct {
		Type          string                                     `json:"_type"`
		PredicateType string                                     `json:"predicateType"`
		Subject       []internalprovenance.Subject               `json:"subject"`
		Predicate     internalprovenance.SLSAProvenancePredicate `json:"predicate"`
	}
	if err := json.Unmarshal(contents, &statement); err != nil {
		t.Fatalf("unmarshal statement: %v", err)
	}

	if got, want := statement.Predicate.RunDetails.Builder.ID, "acme/images/.github/workflows/build.yml@refs/heads/main"; got != want {
		t.Fatalf("unexpected builder id %q, want %q", got, want)
	}
	if got, want := statement.Predicate.RunDetails.Metadata.InvocationID, "run-42"; got != want {
		t.Fatalf("unexpected invocation id %q, want %q", got, want)
	}
	if got, want := statement.Predicate.RunDetails.Metadata.StartedOn, "2026-07-04T10:00:00Z"; got != want {
		t.Fatalf("unexpected startedOn %q, want %q", got, want)
	}
	if got, want := statement.Predicate.RunDetails.Metadata.FinishedOn, "2026-07-04T10:12:00Z"; got != want {
		t.Fatalf("unexpected finishedOn %q, want %q", got, want)
	}

	externalParameters := statement.Predicate.BuildDefinition.ExternalParameters
	if got, want := externalParameters["template"], "ubuntu.pkr.hcl"; got != want {
		t.Fatalf("unexpected template %v, want %q", got, want)
	}
	if got, want := statement.Predicate.BuildDefinition.ResolvedDependencies[0].URI, "git+https://github.com/acme/images@refs/heads/main"; got != want {
		t.Fatalf("unexpected source uri %q, want %q", got, want)
	}
	if got, want := statement.Predicate.BuildDefinition.ResolvedDependencies[0].Digest["gitCommit"], "deadbeef"; got != want {
		t.Fatalf("unexpected source digest %q, want %q", got, want)
	}

	userVariables, ok := externalParameters["userVariables"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected userVariables map, got %T", externalParameters["userVariables"])
	}
	if got, want := userVariables["region"], "us-east-1"; got != want {
		t.Fatalf("unexpected env user variable %v, want %q", got, want)
	}
	if got, want := userVariables["role"], "web"; got != want {
		t.Fatalf("unexpected config user variable %v, want %q", got, want)
	}

	onlyBuilds, ok := externalParameters["onlyBuilds"].([]interface{})
	if !ok || len(onlyBuilds) != 1 || onlyBuilds[0] != "qemu.ubuntu" {
		t.Fatalf("unexpected onlyBuilds value %#v", externalParameters["onlyBuilds"])
	}
}

func TestPostProcessorWritesSBOMAttestation(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	outputDir := t.TempDir()
	config := mustTemplateJSON(t, map[string]interface{}{
		"post-processors": []map[string]interface{}{{
			"type":       "provenance",
			"output_dir": outputDir,
			"sbom":       true,
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var postProcessor PostProcessor
	if err := postProcessor.Configure(tpl.PostProcessors[0][0].Config); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}
	postProcessor.generateSBOM = func(context.Context, internalsbom.Config) ([]byte, error) {
		return []byte(`{"bomFormat":"CycloneDX","specVersion":"1.5"}`), nil
	}

	_, _, _, err = postProcessor.PostProcess(context.Background(), packersdk.TestUi(t), artifact)
	if err != nil {
		t.Fatalf("post-process artifact: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "package.txt.sbom.cdx.json")); err != nil {
		t.Fatalf("expected raw sbom output: %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(outputDir, "package.txt.sbom.att.json"))
	if err != nil {
		t.Fatalf("read sbom attestation: %v", err)
	}

	var statement internalprovenance.Statement
	if err := json.Unmarshal(contents, &statement); err != nil {
		t.Fatalf("unmarshal sbom attestation: %v", err)
	}
	if got, want := statement.PredicateType, "https://cyclonedx.org/bom"; got != want {
		t.Fatalf("unexpected SBOM predicate type %q, want %q", got, want)
	}
}

func TestPostProcessorRegeneratesStaleSBOM(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	outputDir := t.TempDir()
	staleSBOM := []byte(`{"bomFormat":"CycloneDX","specVersion":"1.5","stale":true}`)
	if err := os.WriteFile(filepath.Join(outputDir, "package.txt.sbom.cdx.json"), staleSBOM, 0664); err != nil {
		t.Fatalf("write stale sbom: %v", err)
	}

	config := mustTemplateJSON(t, map[string]interface{}{
		"post-processors": []map[string]interface{}{{
			"type":       "provenance",
			"output_dir": outputDir,
			"sbom":       true,
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var postProcessor PostProcessor
	if err := postProcessor.Configure(tpl.PostProcessors[0][0].Config); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}
	freshSBOM := []byte(`{"bomFormat":"CycloneDX","specVersion":"1.5","fresh":true}`)
	generated := false
	postProcessor.generateSBOM = func(context.Context, internalsbom.Config) ([]byte, error) {
		generated = true
		return freshSBOM, nil
	}

	_, _, _, err = postProcessor.PostProcess(context.Background(), packersdk.TestUi(t), artifact)
	if err != nil {
		t.Fatalf("post-process artifact: %v", err)
	}

	if !generated {
		t.Fatalf("expected SBOM to be regenerated rather than reused")
	}
	if got, want := readFileString(t, filepath.Join(outputDir, "package.txt.sbom.cdx.json")), string(freshSBOM); got != want {
		t.Fatalf("expected stale SBOM to be overwritten with freshly generated contents")
	}

	contents, err := os.ReadFile(filepath.Join(outputDir, "package.txt.sbom.att.json"))
	if err != nil {
		t.Fatalf("read sbom attestation: %v", err)
	}

	var statement internalprovenance.Statement
	if err := json.Unmarshal(contents, &statement); err != nil {
		t.Fatalf("unmarshal sbom attestation: %v", err)
	}
	if got, want := statement.PredicateType, "https://cyclonedx.org/bom"; got != want {
		t.Fatalf("unexpected SBOM predicate type %q, want %q", got, want)
	}
}

func TestPostProcessorSignsAttestationsWithConfiguredVerifier(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	privateKeyPath, publicKeyPath := writeSigningKeypair(t)
	outputDir := t.TempDir()
	config := mustTemplateJSON(t, map[string]interface{}{
		"post-processors": []map[string]interface{}{{
			"type":         "provenance",
			"output_dir":   outputDir,
			"signing_mode": "key",
			"signer":       privateKeyPath,
			"verifier":     publicKeyPath,
			"sbom":         true,
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var postProcessor PostProcessor
	if err := postProcessor.Configure(tpl.PostProcessors[0][0].Config); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}
	postProcessor.generateSBOM = func(context.Context, internalsbom.Config) ([]byte, error) {
		return []byte(`{"bomFormat":"CycloneDX","specVersion":"1.5"}`), nil
	}

	_, _, _, err = postProcessor.PostProcess(context.Background(), packersdk.TestUi(t), artifact)
	if err != nil {
		t.Fatalf("post-process artifact: %v", err)
	}

	assertSignedEnvelope(t, filepath.Join(outputDir, "package.txt.provenance.json"))
	assertSignedEnvelope(t, filepath.Join(outputDir, "package.txt.sbom.att.json"))
	if _, err := os.Stat(filepath.Join(outputDir, "package.txt.sbom.cdx.json")); err != nil {
		t.Fatalf("expected raw sbom output: %v", err)
	}
}

func TestPostProcessorRejectsMismatchedVerifier(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	privateKeyPath, _ := writeSigningKeypair(t)
	_, mismatchedVerifierPath := writeSigningKeypair(t)
	outputDir := t.TempDir()
	config := mustTemplateJSON(t, map[string]interface{}{
		"post-processors": []map[string]interface{}{{
			"type":         "provenance",
			"output_dir":   outputDir,
			"signing_mode": "key",
			"signer":       privateKeyPath,
			"verifier":     mismatchedVerifierPath,
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var postProcessor PostProcessor
	if err := postProcessor.Configure(tpl.PostProcessors[0][0].Config); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}

	_, _, _, err = postProcessor.PostProcess(context.Background(), packersdk.TestUi(t), artifact)
	if err == nil {
		t.Fatalf("expected post-process to fail with mismatched verifier")
	}
}

func TestPostProcessorWritesSigstoreBundleForKeylessAttestations(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	outputDir := t.TempDir()
	config := mustTemplateJSON(t, map[string]interface{}{
		"post-processors": []map[string]interface{}{{
			"type":                "provenance",
			"output_dir":          outputDir,
			"signing_mode":        "keyless",
			"upload_tlog":         false,
			"keyless_identity":    "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main",
			"keyless_oidc_issuer": "https://token.actions.githubusercontent.com",
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	originalBundleBuilder := buildSigstoreBundleForSigner
	buildSigstoreBundleForSigner = func(context.Context, internalattestation.Signer, internalattestation.BackendConfig, string, []byte) (internalattestation.Envelope, []byte, error) {
		return internalattestation.Envelope{
			PayloadType: internalattestation.InTotoPayloadType,
			Payload:     "cGF5bG9hZA==",
			Signatures:  []internalattestation.EnvelopeSignature{{Sig: "c2ln", Cert: "cert"}},
		}, []byte(`{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`), nil
	}
	t.Cleanup(func() { buildSigstoreBundleForSigner = originalBundleBuilder })

	var postProcessor PostProcessor
	if err := postProcessor.Configure(tpl.PostProcessors[0][0].Config); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}
	postProcessor.signingResourcesFn = func(context.Context, internalattestation.BackendConfig) (internalattestation.Signer, internalattestation.Verifier, error) {
		return fakeEnvelopeSigner{}, fakeEnvelopeVerifier{}, nil
	}

	_, _, _, err = postProcessor.PostProcess(context.Background(), packersdk.TestUi(t), artifact)
	if err != nil {
		t.Fatalf("post-process artifact: %v", err)
	}

	assertSignedEnvelope(t, filepath.Join(outputDir, "package.txt.provenance.json"))
	if _, err := os.Stat(filepath.Join(outputDir, "package.txt.provenance.sigstore.json")); err != nil {
		t.Fatalf("expected Sigstore bundle output: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "package.txt.sbom.att.sigstore.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no SBOM bundle sidecar when sbom is disabled, got %v", err)
	}
	if got, want := strings.TrimSpace(readFileString(t, filepath.Join(outputDir, "package.txt.provenance.sigstore.json"))), `{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`; got != want {
		t.Fatalf("unexpected bundle contents %q, want %q", got, want)
	}
	_ = bytes.Buffer{}
}

func TestSigningBackendConfigAcceptsKMSReferences(t *testing.T) {
	postProcessor := PostProcessor{}
	postProcessor.config.SigningMode = internalattestation.SigningModeKMS
	postProcessor.config.Signer = "awskms://alias/example"
	postProcessor.config.Verifier = "keys/provenance-signing.pub.pem"
	postProcessor.env = map[string]string{"AWS_REGION": "us-east-1"}

	backendConfig, err := postProcessor.signingBackendConfig()
	if err != nil {
		t.Fatalf("build signing backend config: %v", err)
	}

	if got, want := backendConfig.Mode, internalattestation.SigningModeKMS; got != want {
		t.Fatalf("unexpected mode %q, want %q", got, want)
	}
	if got, want := backendConfig.SignerRef, "awskms://alias/example"; got != want {
		t.Fatalf("unexpected signer ref %q, want %q", got, want)
	}
	if got, want := backendConfig.VerifierRef, "keys/provenance-signing.pub.pem"; got != want {
		t.Fatalf("unexpected verifier ref %q, want %q", got, want)
	}
	if got, want := backendConfig.Env["AWS_REGION"], "us-east-1"; got != want {
		t.Fatalf("unexpected copied environment %q, want %q", got, want)
	}
}

func TestSigningBackendConfigRejectsUnknownKMSReferences(t *testing.T) {
	postProcessor := PostProcessor{}
	postProcessor.config.SigningMode = internalattestation.SigningModeKMS
	postProcessor.config.Signer = "kms://example"

	_, err := postProcessor.signingBackendConfig()
	if err == nil {
		t.Fatalf("expected unknown KMS reference to be rejected")
	}
	if !strings.Contains(err.Error(), "recognized KMS or Vault URI") {
		t.Fatalf("unexpected KMS validation error: %v", err)
	}
}

func TestSigningBackendConfigIncludesKeylessFulcioURL(t *testing.T) {
	postProcessor := PostProcessor{}
	postProcessor.config.SigningMode = internalattestation.SigningModeKeyless
	postProcessor.config.FulcioURL = "https://fulcio.example.test"
	postProcessor.config.KeylessIdentity = "https://github.com/org/repo/.github/workflows/build.yml@refs/heads/main"
	postProcessor.config.KeylessOIDCIssuer = "https://token.actions.githubusercontent.com"
	postProcessor.env = map[string]string{"SIGSTORE_ID_TOKEN": "token"}

	backendConfig, err := postProcessor.signingBackendConfig()
	if err != nil {
		t.Fatalf("build signing backend config: %v", err)
	}

	if got, want := backendConfig.Mode, internalattestation.SigningModeKeyless; got != want {
		t.Fatalf("unexpected mode %q, want %q", got, want)
	}
	if got, want := backendConfig.FulcioURL, "https://fulcio.example.test"; got != want {
		t.Fatalf("unexpected Fulcio URL %q, want %q", got, want)
	}
	if backendConfig.VerifierRef != "" {
		t.Fatalf("keyless backend config must not carry a verifier ref, got %q", backendConfig.VerifierRef)
	}
	if got, want := backendConfig.Env["SIGSTORE_ID_TOKEN"], "token"; got != want {
		t.Fatalf("unexpected copied environment %q, want %q", got, want)
	}
}

func TestSigningBackendConfigRejectsKeylessVerifierOverride(t *testing.T) {
	postProcessor := PostProcessor{}
	postProcessor.config.SigningMode = internalattestation.SigningModeKeyless
	postProcessor.config.Verifier = "keys/provenance-signing.pub.pem"
	postProcessor.config.KeylessIdentity = "https://github.com/org/repo/.github/workflows/build.yml@refs/heads/main"
	postProcessor.config.KeylessOIDCIssuer = "https://token.actions.githubusercontent.com"
	postProcessor.env = map[string]string{"SIGSTORE_ID_TOKEN": "token"}

	_, err := postProcessor.signingBackendConfig()
	if err == nil {
		t.Fatalf("expected keyless config with verifier override to fail")
	}
	if !strings.Contains(err.Error(), "verifier overrides") {
		t.Fatalf("unexpected keyless verifier override error: %v", err)
	}
}

func TestSigningBackendConfigRejectsKeylessWithoutIdentityPolicy(t *testing.T) {
	postProcessor := PostProcessor{}
	postProcessor.config.SigningMode = internalattestation.SigningModeKeyless
	postProcessor.config.FulcioURL = "https://fulcio.example.test"
	postProcessor.env = map[string]string{"SIGSTORE_ID_TOKEN": "token"}

	_, err := postProcessor.signingBackendConfig()
	if err == nil {
		t.Fatalf("expected keyless config without identity policy to fail")
	}
	if !strings.Contains(err.Error(), "keyless_identity") {
		t.Fatalf("unexpected keyless validation error: %v", err)
	}
}

func TestSigningBackendConfigIncludesKeylessPolicy(t *testing.T) {
	postProcessor := PostProcessor{}
	postProcessor.config.SigningMode = internalattestation.SigningModeKeyless
	postProcessor.config.FulcioURL = "https://fulcio.example.test"
	postProcessor.config.KeylessIdentity = "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"
	postProcessor.config.KeylessOIDCIssuer = "https://token.actions.githubusercontent.com"
	postProcessor.config.TrustedRootPath = "testdata/trusted-root.json"
	postProcessor.env = map[string]string{"SIGSTORE_ID_TOKEN": "token"}

	backendConfig, err := postProcessor.signingBackendConfig()
	if err != nil {
		t.Fatalf("build keyless signing backend config: %v", err)
	}

	if got, want := backendConfig.KeylessIdentity, "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"; got != want {
		t.Fatalf("unexpected keyless identity %q, want %q", got, want)
	}
	if got, want := backendConfig.KeylessOIDCIssuer, "https://token.actions.githubusercontent.com"; got != want {
		t.Fatalf("unexpected keyless OIDC issuer %q, want %q", got, want)
	}
	if got, want := backendConfig.TrustedRootPath, "testdata/trusted-root.json"; got != want {
		t.Fatalf("unexpected trusted root path %q, want %q", got, want)
	}
}

func TestSigningBackendConfigIncludesKeylessRekorSettings(t *testing.T) {
	postProcessor := PostProcessor{}
	postProcessor.config.SigningMode = internalattestation.SigningModeKeyless
	postProcessor.config.FulcioURL = "https://fulcio.example.test"
	postProcessor.config.RekorURL = "https://rekor.example.test"
	postProcessor.config.UploadTlog = true
	postProcessor.config.KeylessIdentity = "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"
	postProcessor.config.KeylessOIDCIssuer = "https://token.actions.githubusercontent.com"
	postProcessor.env = map[string]string{"SIGSTORE_ID_TOKEN": "token"}

	backendConfig, err := postProcessor.signingBackendConfig()
	if err != nil {
		t.Fatalf("build keyless signing backend config: %v", err)
	}

	if got, want := backendConfig.RekorURL, "https://rekor.example.test"; got != want {
		t.Fatalf("unexpected Rekor URL %q, want %q", got, want)
	}
	if !backendConfig.UploadTlog {
		t.Fatalf("expected upload_tlog to be enabled")
	}
}

func assertSignedEnvelope(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read signed envelope %q: %v", path, err)
	}

	var envelope internalattestation.Envelope
	if err := json.Unmarshal(contents, &envelope); err != nil {
		t.Fatalf("unmarshal envelope %q: %v", path, err)
	}
	if got, want := envelope.PayloadType, internalattestation.InTotoPayloadType; got != want {
		t.Fatalf("unexpected payload type %q, want %q", got, want)
	}
	if len(envelope.Signatures) != 1 {
		t.Fatalf("expected exactly one signature in %q", path)
	}
}

func writeSigningKeypair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	privateKeyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}

	publicKeyDER, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	dir := t.TempDir()
	privateKeyPath := filepath.Join(dir, "signer.pem")
	publicKeyPath := filepath.Join(dir, "verifier.pem")

	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyDER}), 0600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	if err := os.WriteFile(publicKeyPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER}), 0600); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	return privateKeyPath, publicKeyPath
}

type fakeEnvelopeSigner struct{}

func (fakeEnvelopeSigner) Sign(context.Context, string, []byte) (internalattestation.Signature, error) {
	return internalattestation.Signature{Sig: []byte("sig")}, nil
}

func (fakeEnvelopeSigner) Verifier(context.Context, internalattestation.BackendConfig) (internalattestation.Verifier, error) {
	return fakeEnvelopeVerifier{}, nil
}

type fakeEnvelopeVerifier struct{}

func (fakeEnvelopeVerifier) Verify(context.Context, string, []byte, []byte) error {
	return nil
}

func (fakeEnvelopeVerifier) KeyID() string {
	return ""
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}
	return string(contents)
}

func TestOutputStemDisambiguatesByBuildName(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	var pp PostProcessor

	// Without a build name the stem is just the artifact's base filename.
	if got, want := pp.outputStem(artifact), "package.txt"; got != want {
		t.Fatalf("without build name: got %q, want %q", got, want)
	}

	// A build name is prefixed so parallel builds sharing an output_dir do not
	// collide on the same output paths.
	pp.config.PackerBuildName = "amazon-ebs.linux"
	if got, want := pp.outputStem(artifact), "amazon-ebs.linux.package.txt"; got != want {
		t.Fatalf("with build name: got %q, want %q", got, want)
	}

	paths, err := pp.outputPaths(artifact)
	if err != nil {
		t.Fatalf("output paths: %v", err)
	}
	if got, want := filepath.Base(paths.ProvenanceStatement), "amazon-ebs.linux.package.txt.provenance.json"; got != want {
		t.Fatalf("unexpected provenance path %q, want %q", got, want)
	}
}

func TestAtomicWriteFileReplacesExistingWithoutLeftovers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	if err := os.WriteFile(path, []byte("stale"), 0644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := atomicWriteFile(path, []byte("fresh-content"), 0664); err != nil {
		t.Fatalf("atomic write: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(got) != "fresh-content" {
		t.Fatalf("unexpected contents %q", string(got))
	}

	// No temporary files should be left behind.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "out.json" {
		t.Fatalf("expected only out.json in directory, found %v", entries)
	}
}

func buildFileArtifact(t *testing.T) packersdk.Artifact {
	t.Helper()

	target := filepath.Join(t.TempDir(), "package.txt")
	config := mustTemplateJSON(t, map[string]interface{}{
		"builders": []map[string]string{{
			"type":    "file",
			"target":  target,
			"content": "Hello world!",
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var builder filebuilder.Builder
	_, warnings, err := builder.Prepare(tpl.Builders["file"].Config)
	if err != nil {
		t.Fatalf("prepare builder: %v", err)
	}
	if len(warnings) > 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	artifact, err := builder.Run(context.Background(), packersdk.TestUi(t), nil)
	if err != nil {
		t.Fatalf("run builder: %v", err)
	}

	return artifact
}

func mustTemplateJSON(t *testing.T, value interface{}) string {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal template config: %v", err)
	}

	return string(encoded)
}
