// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"strings"

	internalattestation "github.com/hashicorp/packer/internal/attestation"
	"github.com/posener/complete"
)

type VerifyAttestationCommand struct {
	Meta
}

type VerifyAttestationArgs struct {
	AttestationPath          string
	SigningMode              string
	Key                      string
	Verifier                 string
	PredicateType            string
	BuilderID                string
	SourceURI                string
	ArtifactPath             string
	TrustedRootPath          string
	KeylessIdentity          string
	KeylessOIDCIssuer        string
	SigstoreBundlePath       string
	RequireTransparencyLog   bool
	RequireObserverTimestamp bool
}

func (c *VerifyAttestationCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *VerifyAttestationCommand) ParseArgs(args []string) (*VerifyAttestationArgs, int) {
	var cfg VerifyAttestationArgs

	flags := c.Meta.FlagSet("verify-attestation")
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	flags.StringVar(&cfg.SigningMode, "signing-mode", "", "Signing mode used for the attestation")
	flags.StringVar(&cfg.Key, "key", "", "PEM path or KMS/Vault URI")
	flags.StringVar(&cfg.Verifier, "verifier", "", "PEM verifier path")
	flags.StringVar(&cfg.PredicateType, "predicate-type", "", "Expected attestation predicate type")
	flags.StringVar(&cfg.BuilderID, "builder-id", "", "Expected SLSA builder ID")
	flags.StringVar(&cfg.SourceURI, "source-uri", "", "Expected resolved source URI")
	flags.StringVar(&cfg.ArtifactPath, "artifact", "", "Artifact path to match against attestation subjects")
	flags.StringVar(&cfg.TrustedRootPath, "trusted-root-path", "", "Optional Sigstore trusted-root JSON file")
	flags.StringVar(&cfg.KeylessIdentity, "keyless-identity", "", "Expected keyless signing identity")
	flags.StringVar(&cfg.KeylessOIDCIssuer, "keyless-oidc-issuer", "", "Expected keyless OIDC issuer")
	flags.StringVar(&cfg.SigstoreBundlePath, "bundle", "", "Optional Sigstore bundle JSON file for Rekor or timestamp verification")
	flags.BoolVar(&cfg.RequireTransparencyLog, "require-rekor", false, "Require Rekor transparency log verification using the Sigstore bundle")
	flags.BoolVar(&cfg.RequireObserverTimestamp, "require-timestamp", false, "Require a trusted observer timestamp from Rekor integrated time or RFC3161 timestamp evidence")
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return &cfg, 1
	}
	cfg.AttestationPath = args[0]
	return &cfg, 0
}

func (c *VerifyAttestationCommand) RunContext(ctx context.Context, cfg *VerifyAttestationArgs) int {
	_, err := internalattestation.VerifyAttestationFile(ctx, cfg.AttestationPath, internalattestation.BackendConfig{
		Mode:              cfg.SigningMode,
		SignerRef:         cfg.Key,
		VerifierRef:       cfg.Verifier,
		TrustedRootPath:   cfg.TrustedRootPath,
		KeylessIdentity:   cfg.KeylessIdentity,
		KeylessOIDCIssuer: cfg.KeylessOIDCIssuer,
	}, internalattestation.VerificationPolicy{
		PredicateType:            cfg.PredicateType,
		BuilderID:                cfg.BuilderID,
		SourceURI:                cfg.SourceURI,
		ArtifactPath:             cfg.ArtifactPath,
		SigstoreBundlePath:       cfg.SigstoreBundlePath,
		RequireTransparencyLog:   cfg.RequireTransparencyLog,
		RequireObserverTimestamp: cfg.RequireObserverTimestamp,
	})
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	c.Ui.Say("Attestation verified.")
	return 0
}

func (*VerifyAttestationCommand) Help() string {
	helpText := `
Usage: packer verify-attestation [options] ATTESTATION

  Verifies a signed DSSE attestation and optionally enforces policy checks
  such as predicate type, builder identity, source URI, and subject digest.

Options:

  -signing-mode=MODE            Signing mode: key, kms, keyless. Auto-detected when possible.
  -key=PATH_OR_URI              PEM key path or KMS/Vault URI used for verification when no verifier is supplied.
  -verifier=PATH                PEM verifier path.
  -predicate-type=TYPE          Expected attestation predicate type.
  -builder-id=ID                Expected SLSA builder ID.
  -source-uri=URI               Expected resolved source URI.
  -artifact=PATH                Artifact path to match against attestation subjects.
  -trusted-root-path=PATH       Optional Sigstore trusted-root JSON for keyless verification.
  -keyless-identity=IDENTITY    Expected keyless signing identity.
  -keyless-oidc-issuer=ISSUER   Expected keyless OIDC issuer.
	-bundle=PATH                  Optional Sigstore bundle JSON for Rekor or timestamp verification.
	-require-rekor                Require Rekor transparency log verification from the bundle.
	-require-timestamp            Require a trusted observer timestamp from Rekor integrated time or RFC3161 evidence.
`

	return strings.TrimSpace(helpText)
}

func (*VerifyAttestationCommand) Synopsis() string {
	return "verify a signed attestation against policy"
}

func (*VerifyAttestationCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*VerifyAttestationCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-signing-mode":        complete.PredictNothing,
		"-key":                 complete.PredictNothing,
		"-verifier":            complete.PredictNothing,
		"-predicate-type":      complete.PredictNothing,
		"-builder-id":          complete.PredictNothing,
		"-source-uri":          complete.PredictNothing,
		"-artifact":            complete.PredictNothing,
		"-trusted-root-path":   complete.PredictNothing,
		"-keyless-identity":    complete.PredictNothing,
		"-keyless-oidc-issuer": complete.PredictNothing,
		"-bundle":              complete.PredictNothing,
		"-require-rekor":       complete.PredictNothing,
		"-require-timestamp":   complete.PredictNothing,
	}
}
