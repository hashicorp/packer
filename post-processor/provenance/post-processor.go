// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package provenance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	internalattestation "github.com/hashicorp/packer/internal/attestation"
	internalprovenance "github.com/hashicorp/packer/internal/provenance"
	internalsbom "github.com/hashicorp/packer/internal/sbom"
)

var buildSigstoreBundleForSigner = internalattestation.BuildBundleForSigner

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Provenance        config.Trilean    `mapstructure:"provenance"`
	BuildType         string            `mapstructure:"build_type"`
	OutputDir         string            `mapstructure:"output_dir"`
	TemplatePath      string            `mapstructure:"template"`
	OnlyBuilds        []string          `mapstructure:"only_builds"`
	UserVariables     map[string]string `mapstructure:"user_variables"`
	SourceURI         string            `mapstructure:"source_uri"`
	SBOM              bool              `mapstructure:"sbom"`
	SBOMFormat        string            `mapstructure:"sbom_format"`
	SBOMScanPath      string            `mapstructure:"sbom_scan_path"`
	SBOMScope         string            `mapstructure:"sbom_scope"`
	SBOMExclude       []string          `mapstructure:"sbom_exclude"`
	SigningMode       string            `mapstructure:"signing_mode"`
	Signer            string            `mapstructure:"signer"`
	Key               string            `mapstructure:"key"`
	Verifier          string            `mapstructure:"verifier"`
	FulcioURL         string            `mapstructure:"fulcio_url"`
	RekorURL          string            `mapstructure:"rekor_url"`
	UploadTlog        bool              `mapstructure:"upload_tlog"`
	TrustedRootPath   string            `mapstructure:"trusted_root_path"`
	KeylessIdentity   string            `mapstructure:"keyless_identity"`
	KeylessOIDCIssuer string            `mapstructure:"keyless_oidc_issuer"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config             Config
	now                func() time.Time
	env                map[string]string
	workingDir         string
	generateSBOM       func(context.Context, internalsbom.Config) ([]byte, error)
	signingResourcesFn func(context.Context, internalattestation.BackendConfig) (internalattestation.Signer, internalattestation.Verifier, error)
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "packer.post-processor.provenance",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Defaults are applied after decoding because the HCL2 decode path zeroes
	// unset fields, which would otherwise clobber any pre-decode defaults.
	if p.config.BuildType == "" {
		p.config.BuildType = internalprovenance.DefaultBuildType
	}
	if p.config.SBOMFormat == "" {
		p.config.SBOMFormat = string(internalsbom.FormatCycloneDX)
	}
	if p.config.SBOMScope == "" {
		p.config.SBOMScope = internalsbom.ScopeSquashed
	}
	if p.config.SigningMode == "" {
		p.config.SigningMode = internalattestation.SigningModeNone
	}
	if p.config.FulcioURL == "" {
		p.config.FulcioURL = "https://fulcio.sigstore.dev"
	}
	if p.config.RekorURL == "" {
		p.config.RekorURL = "https://rekor.sigstore.dev"
	}

	if p.config.OutputDir != "" {
		if err := interpolate.Validate(p.config.OutputDir, &p.config.ctx); err != nil {
			return fmt.Errorf("error parsing output_dir template: %w", err)
		}
	}

	if _, err := p.signingBackendConfig(); err != nil {
		return err
	}

	if p.generateSBOM == nil {
		p.generateSBOM = func(ctx context.Context, cfg internalsbom.Config) ([]byte, error) {
			return internalsbom.NewGenerator(cfg).Generate(ctx)
		}
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, source packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	if p.config.Provenance.False() {
		return source, true, true, nil
	}

	startedAt := p.currentTime().UTC()
	env := p.currentEnv()

	select {
	case <-ctx.Done():
		return source, true, true, ctx.Err()
	default:
	}

	subjects, err := internalprovenance.DeriveSubjects(source)
	if err != nil {
		return source, true, true, err
	}

	var byproducts []internalprovenance.Byproduct
	if len(source.Files()) == 0 {
		identityRecord, err := internalprovenance.DeriveIdentityRecord(source)
		if err != nil {
			return source, true, true, err
		}

		byproducts = append(byproducts, internalprovenance.Byproduct{
			Name:    "cloud-artifact-identity",
			Content: identityRecord,
		})
	}

	invocationID := internalprovenance.DetectInvocationID(env)
	if invocationID == "" {
		invocationID, _ = uuid.GenerateUUID()
	}

	finishedAt := p.currentTime().UTC()

	predicate := internalprovenance.BuildSLSAPredicate(internalprovenance.PredicateInput{
		BuildType:            p.config.BuildType,
		ExternalParameters:   p.externalParameters(env),
		InternalParameters:   p.internalParameters(),
		ResolvedDependencies: p.resolvedDependencies(env),
		BuilderID:            internalprovenance.DetectBuilderID(env),
		Byproducts:           byproducts,
		InvocationID:         invocationID,
		StartedOn:            startedAt.Format(time.RFC3339),
		FinishedOn:           finishedAt.Format(time.RFC3339),
	})
	statement := internalprovenance.WrapInToto(subjects, internalprovenance.SLSAProvenanceV1PredicateType, predicate)

	paths, err := p.outputPaths(source)
	if err != nil {
		return source, true, true, err
	}

	if err := p.writeAttestation(ctx, ui, statement, paths.ProvenanceStatement); err != nil {
		return source, true, true, err
	}

	if p.config.SBOM {
		if err := p.writeSBOMAttestation(ctx, ui, source, subjects, paths); err != nil {
			return source, true, true, err
		}
	}

	return source, true, true, nil
}

const (
	predicateTypeCycloneDX = "https://cyclonedx.org/bom"
	predicateTypeSPDX      = "https://spdx.dev/Document"
)

// redactedSensitiveValue replaces sensitive user-variable values in the
// provenance predicate so secrets are never written to the attestation.
const redactedSensitiveValue = "[sensitive value redacted]"

type outputPaths struct {
	BaseDir             string
	Stem                string
	ProvenanceStatement string
	SBOMRaw             string
	SBOMAttestation     string
}

func (p *PostProcessor) writeSBOMAttestation(ctx context.Context, ui packersdk.Ui, source packersdk.Artifact, subjects []internalprovenance.Subject, paths outputPaths) error {
	format, rawSBOM, err := p.resolveSBOM(ctx, source, paths)
	if err != nil {
		return err
	}

	predicate, predicateType, err := buildSBOMPredicate(rawSBOM, format)
	if err != nil {
		return err
	}

	statement := internalprovenance.WrapInToto(subjects, predicateType, predicate)
	if err := p.writeAttestation(ctx, ui, statement, paths.SBOMAttestation); err != nil {
		return err
	}

	ui.Say(fmt.Sprintf("Wrote SBOM to %s", paths.SBOMRaw))
	return nil
}

func (p *PostProcessor) writeAttestation(ctx context.Context, ui packersdk.Ui, statement interface{}, outputPath string) error {
	if p.config.SigningMode == internalattestation.SigningModeNone {
		payload, err := json.MarshalIndent(statement, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal attestation payload: %w", err)
		}

		if err := atomicWriteFile(outputPath, payload, 0664); err != nil {
			return fmt.Errorf("write attestation %q: %w", outputPath, err)
		}

		ui.Say(fmt.Sprintf("Wrote attestation to %s", outputPath))
		return nil
	}

	backendConfig, err := p.signingBackendConfig()
	if err != nil {
		return err
	}

	signer, verifier, err := p.signingResources(ctx, backendConfig)
	if err != nil {
		return err
	}

	payload, err := internalattestation.MarshalPayload(statement)
	if err != nil {
		return fmt.Errorf("marshal canonical attestation payload: %w", err)
	}

	bundlePath := sigstoreBundleOutputPath(outputPath)
	bundleJSON := []byte(nil)
	var envelope internalattestation.Envelope
	if backendConfig.Mode == internalattestation.SigningModeKeyless {
		envelope, bundleJSON, err = buildSigstoreBundleForSigner(ctx, signer, backendConfig, internalattestation.InTotoPayloadType, payload)
		if err != nil {
			return fmt.Errorf("sign attestation with Sigstore bundle: %w", err)
		}
	} else {
		signature, signErr := signer.Sign(ctx, internalattestation.InTotoPayloadType, payload)
		if signErr != nil {
			return fmt.Errorf("sign attestation: %w", signErr)
		}
		envelope = internalattestation.NewEnvelope(internalattestation.InTotoPayloadType, payload, signature)
	}

	if err := internalattestation.VerifyEnvelope(ctx, envelope, verifier); err != nil {
		return fmt.Errorf("verify signed attestation: %w", err)
	}

	output, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal signed envelope: %w", err)
	}

	if err := atomicWriteFile(outputPath, output, 0664); err != nil {
		return fmt.Errorf("write attestation %q: %w", outputPath, err)
	}

	if len(bundleJSON) > 0 {
		if err := atomicWriteFile(bundlePath, bundleJSON, 0664); err != nil {
			return fmt.Errorf("write Sigstore bundle %q: %w", bundlePath, err)
		}
		ui.Say(fmt.Sprintf("Wrote Sigstore bundle to %s", bundlePath))
	}

	ui.Say(fmt.Sprintf("Wrote attestation to %s", outputPath))
	return nil
}

func (p *PostProcessor) signingResources(ctx context.Context, backendConfig internalattestation.BackendConfig) (internalattestation.Signer, internalattestation.Verifier, error) {
	if p.signingResourcesFn != nil {
		return p.signingResourcesFn(ctx, backendConfig)
	}

	if backendConfig.Mode == internalattestation.SigningModeNone {
		return nil, nil, nil
	}

	signer, err := internalattestation.NewSigner(ctx, backendConfig)
	if err != nil {
		return nil, nil, err
	}

	verifier, err := internalattestation.NewVerifier(ctx, backendConfig, signer)
	if err != nil {
		return nil, nil, err
	}

	return signer, verifier, nil
}

func (p *PostProcessor) signingBackendConfig() (internalattestation.BackendConfig, error) {
	mode := p.config.SigningMode
	if mode == "" {
		mode = internalattestation.SigningModeNone
	}

	signerRef := p.config.Signer
	if p.config.Key != "" {
		if signerRef != "" && signerRef != p.config.Key {
			return internalattestation.BackendConfig{}, fmt.Errorf("signer and key must match when both are set")
		}
		signerRef = p.config.Key
	}

	switch mode {
	case internalattestation.SigningModeNone:
		return internalattestation.BackendConfig{Mode: mode}, nil
	case internalattestation.SigningModeKey:
		if signerRef == "" {
			return internalattestation.BackendConfig{}, fmt.Errorf("signing_mode %q requires signer or key", mode)
		}
		return internalattestation.BackendConfig{
			Mode:        mode,
			SignerRef:   signerRef,
			VerifierRef: p.config.Verifier,
			Env:         p.currentEnv(),
		}, nil
	case internalattestation.SigningModeKMS:
		if signerRef == "" {
			return internalattestation.BackendConfig{}, fmt.Errorf("signing_mode %q requires signer or key", mode)
		}
		if !isRecognizedKMSSigner(signerRef) {
			return internalattestation.BackendConfig{}, fmt.Errorf("signing_mode %q requires a recognized KMS or Vault URI: awskms://, gcpkms://, azurekms://, or hashivault://", mode)
		}
		return internalattestation.BackendConfig{
			Mode:        mode,
			SignerRef:   signerRef,
			VerifierRef: p.config.Verifier,
			Env:         p.currentEnv(),
		}, nil
	case internalattestation.SigningModeKeyless:
		if p.config.Verifier != "" {
			return internalattestation.BackendConfig{}, fmt.Errorf("signing_mode %q does not support verifier overrides; keyless attestations are verified against keyless_identity and keyless_oidc_issuer", mode)
		}
		if strings.TrimSpace(p.config.KeylessIdentity) == "" || strings.TrimSpace(p.config.KeylessOIDCIssuer) == "" {
			return internalattestation.BackendConfig{}, fmt.Errorf("signing_mode %q requires keyless_identity and keyless_oidc_issuer", mode)
		}
		return internalattestation.BackendConfig{
			Mode:              mode,
			Env:               p.currentEnv(),
			FulcioURL:         p.config.FulcioURL,
			RekorURL:          p.config.RekorURL,
			UploadTlog:        p.config.UploadTlog,
			TrustedRootPath:   p.config.TrustedRootPath,
			KeylessIdentity:   p.config.KeylessIdentity,
			KeylessOIDCIssuer: p.config.KeylessOIDCIssuer,
		}, nil
	default:
		return internalattestation.BackendConfig{}, fmt.Errorf("signing_mode %q is not implemented", mode)
	}
}

func isRecognizedKMSSigner(value string) bool {
	for _, prefix := range []string{"awskms://", "gcpkms://", "azurekms://", "hashivault://"} {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}

	return false
}

func (p *PostProcessor) resolveSBOM(ctx context.Context, source packersdk.Artifact, paths outputPaths) (internalsbom.Format, []byte, error) {
	// The SBOM is always regenerated so it reflects the artifact being attested.
	// Reusing a pre-existing SBOM file could attest stale contents if the
	// artifact changed between runs.
	format, err := internalsbom.ParseFormatFromArgs(p.config.SBOMFormat)
	if err != nil {
		return "", nil, err
	}

	scanPath, err := p.resolveSBOMScanPath(source)
	if err != nil {
		return "", nil, err
	}

	rawSBOM, err := p.generateSBOM(ctx, internalsbom.Config{
		ScanPath: scanPath,
		Format:   format,
		Scope:    p.config.SBOMScope,
		Exclude:  append([]string(nil), p.config.SBOMExclude...),
	})
	if err != nil {
		return "", nil, fmt.Errorf("generate SBOM: %w", err)
	}

	if err := atomicWriteFile(paths.SBOMRaw, rawSBOM, 0664); err != nil {
		return "", nil, fmt.Errorf("write SBOM %q: %w", paths.SBOMRaw, err)
	}

	return format, rawSBOM, nil
}

func (p *PostProcessor) resolveSBOMScanPath(source packersdk.Artifact) (string, error) {
	if p.config.SBOMScanPath != "" {
		return p.config.SBOMScanPath, nil
	}

	files := source.Files()
	if len(files) == 1 {
		return files[0], nil
	}
	if len(files) > 1 {
		parent := filepath.Dir(files[0])
		for _, file := range files[1:] {
			if filepath.Dir(file) != parent {
				return "", fmt.Errorf("sbom=true requires sbom_scan_path when artifact files span multiple directories")
			}
		}
		return parent, nil
	}

	return "", fmt.Errorf("sbom=true requires local artifact files or sbom_scan_path")
}

func buildSBOMPredicate(rawSBOM []byte, format internalsbom.Format) (interface{}, string, error) {
	decoder := json.NewDecoder(strings.NewReader(string(rawSBOM)))
	decoder.UseNumber()

	var predicate interface{}
	if err := decoder.Decode(&predicate); err != nil {
		return nil, "", fmt.Errorf("decode SBOM payload: %w", err)
	}

	switch format {
	case internalsbom.FormatCycloneDX:
		return predicate, predicateTypeCycloneDX, nil
	case internalsbom.FormatSPDX:
		return predicate, predicateTypeSPDX, nil
	default:
		return nil, "", fmt.Errorf("unsupported SBOM format %q", format)
	}
}

func (p *PostProcessor) externalParameters(env map[string]string) map[string]interface{} {
	externalParameters := map[string]interface{}{}

	if p.config.TemplatePath != "" {
		externalParameters["template"] = p.config.TemplatePath
	}
	if len(p.config.OnlyBuilds) > 0 {
		externalParameters["onlyBuilds"] = append([]string(nil), p.config.OnlyBuilds...)
	}

	userVariables := collectUserVariables(env)
	for key, value := range p.config.UserVariables {
		userVariables[key] = value
	}
	redactSensitiveVariables(userVariables, p.config.PackerSensitiveVars)
	if len(userVariables) > 0 {
		externalParameters["userVariables"] = userVariables
	}

	if len(externalParameters) == 0 {
		return nil
	}

	return externalParameters
}

func (p *PostProcessor) internalParameters() map[string]interface{} {
	return map[string]interface{}{
		"packerBuildName":   p.config.PackerBuildName,
		"packerBuilderType": p.config.PackerBuilderType,
	}
}

func (p *PostProcessor) resolvedDependencies(env map[string]string) []internalprovenance.ResolvedDependency {
	workingDir := p.currentWorkingDir()
	dependency, ok := internalprovenance.DetectGitDependency(workingDir, env)
	if p.config.SourceURI != "" {
		if ok {
			dependency.URI = p.config.SourceURI
		} else {
			dependency = internalprovenance.ResolvedDependency{URI: p.config.SourceURI}
			ok = true
		}
	}
	if !ok {
		return nil
	}

	return []internalprovenance.ResolvedDependency{dependency}
}

func (p *PostProcessor) currentEnv() map[string]string {
	if p.env != nil {
		copiedEnv := make(map[string]string, len(p.env))
		for key, value := range p.env {
			copiedEnv[key] = value
		}
		return copiedEnv
	}

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

func (p *PostProcessor) currentTime() time.Time {
	if p.now != nil {
		return p.now()
	}

	return time.Now()
}

func (p *PostProcessor) currentWorkingDir() string {
	if p.workingDir != "" {
		return p.workingDir
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	return workingDir
}

func collectUserVariables(env map[string]string) map[string]string {
	userVariables := map[string]string{}
	keys := make([]string, 0)
	for key := range env {
		if strings.HasPrefix(key, "PKR_VAR_") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		userVariables[strings.TrimPrefix(key, "PKR_VAR_")] = env[key]
	}

	return userVariables
}

// redactSensitiveVariables replaces the values of any user variables whose names
// were marked sensitive (packer_sensitive_variables) so that secrets are not
// embedded in the provenance predicate, per SLSA guidance.
func redactSensitiveVariables(userVariables map[string]string, sensitiveKeys []string) {
	for _, key := range sensitiveKeys {
		if _, ok := userVariables[key]; ok {
			userVariables[key] = redactedSensitiveValue
		}
	}
}

func (p *PostProcessor) outputPaths(source packersdk.Artifact) (outputPaths, error) {
	baseDir := p.config.OutputDir
	if baseDir == "" && len(source.Files()) > 0 {
		baseDir = filepath.Dir(source.Files()[0])
	}
	if baseDir == "" {
		baseDir = "."
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return outputPaths{}, fmt.Errorf("create output dir %q: %w", baseDir, err)
	}

	name := p.outputStem(source)
	sbomFormat := internalsbom.FormatCycloneDX
	if parsed, err := internalsbom.ParseFormatFromArgs(p.config.SBOMFormat); err == nil {
		sbomFormat = parsed
	}
	sbomRaw := filepath.Join(baseDir, name+".sbom.cdx.json")
	if sbomFormat == internalsbom.FormatSPDX {
		sbomRaw = filepath.Join(baseDir, name+".sbom.spdx.json")
	}

	return outputPaths{
		BaseDir:             baseDir,
		Stem:                name,
		ProvenanceStatement: filepath.Join(baseDir, name+".provenance.json"),
		SBOMRaw:             sbomRaw,
		SBOMAttestation:     filepath.Join(baseDir, name+".sbom.att.json"),
	}, nil
}

// outputStem returns the base filename used for all provenance outputs. When a
// build name is available it is prefixed so that parallel builds writing to a
// shared output directory cannot collide on the same output paths.
func (p *PostProcessor) outputStem(source packersdk.Artifact) string {
	base := artifactStem(source)

	buildName := sanitizeFilename(strings.TrimSpace(p.config.PackerBuildName))
	if buildName == "" || base == buildName || strings.HasPrefix(base, buildName+".") {
		return base
	}

	return buildName + "." + base
}

func artifactStem(source packersdk.Artifact) string {
	if files := source.Files(); len(files) > 0 {
		return filepath.Base(files[0])
	}

	return sanitizeFilename(fmt.Sprintf("%s-%s", source.BuilderId(), source.Id()))
}

func sigstoreBundleOutputPath(attestationPath string) string {
	if strings.HasSuffix(attestationPath, ".json") {
		return strings.TrimSuffix(attestationPath, ".json") + ".sigstore.json"
	}

	return attestationPath + ".sigstore.json"
}

func sanitizeFilename(value string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		case r == '.', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, value)
}

// atomicWriteFile writes data to path atomically by writing to a temporary file
// in the same directory and renaming it into place. This prevents partially
// written or interleaved outputs when builds run in parallel, and ensures a
// crash mid-write cannot leave a corrupt attestation on disk.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	committed = true

	return nil
}
