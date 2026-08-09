// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// TestConfigureHCL2DefaultProvenanceEnabled reproduces the HCL2 decode path used
// by packer core (hcldec.Decode -> cty.Value -> Configure) to ensure the
// default-true `provenance` gate survives when the field is omitted in HCL.
func TestConfigureHCL2DefaultProvenanceEnabled(t *testing.T) {
	src := `output_dir = "out"`
	body, diags := hclsyntax.ParseConfig([]byte(src), "test.pkr.hcl", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		t.Fatalf("parse hcl: %s", diags)
	}

	var pp PostProcessor
	spec := pp.ConfigSpec()

	val, decodeDiags := hcldec.Decode(body.Body, spec, nil)
	if decodeDiags.HasErrors() {
		t.Fatalf("decode hcl2 spec: %s", decodeDiags)
	}

	if err := pp.Configure(val); err != nil {
		t.Fatalf("configure post-processor: %v", err)
	}

	if pp.config.Provenance.False() {
		t.Fatalf("expected provenance to remain enabled by default under HCL2 decode")
	}
	if pp.config.SBOMFormat == "" {
		t.Fatalf("expected SBOMFormat default to survive HCL2 decode, got empty")
	}
	if pp.config.SBOMScope == "" {
		t.Fatalf("expected SBOMScope default to survive HCL2 decode, got empty")
	}
	if pp.config.BuildType == "" {
		t.Fatalf("expected BuildType default to survive HCL2 decode, got empty")
	}
	if pp.config.SigningMode == "" {
		t.Fatalf("expected SigningMode default to survive HCL2 decode, got empty")
	}
	if pp.config.FulcioURL == "" {
		t.Fatalf("expected FulcioURL default to survive HCL2 decode, got empty")
	}
	if pp.config.RekorURL == "" {
		t.Fatalf("expected RekorURL default to survive HCL2 decode, got empty")
	}
}
