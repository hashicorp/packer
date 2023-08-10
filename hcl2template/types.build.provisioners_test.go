// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
)

func TestPackerConfig_ParseProvisionerBlock(t *testing.T) {
	tests := []struct {
		name                 string
		inputFile            string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			"success - provisioner is valid",
			"fixtures/well_formed_provisioner.pkr.hcl",
			false,
			"",
		},
		{
			"failure - provisioner override is malformed",
			"fixtures/malformed_override.pkr.hcl",
			true,
			"provisioner's override block must be an HCL object",
		},
		{
			"failure - provisioner override.test is malformed",
			"fixtures/malformed_override_innards.pkr.hcl",
			true,
			"provisioner's override.'test' block must be an HCL object",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := PackerConfig{parser: getBasicParser()}
			f, diags := cfg.parser.ParseHCLFile(test.inputFile)
			if diags.HasErrors() {
				t.Errorf("failed to parse input file %s", test.inputFile)
				for _, d := range diags {
					t.Errorf("%s", d)
				}
				return
			}
			provBlock := f.OutermostBlockAtPos(hcl.Pos{
				Line:   1,
				Column: 1,
				Byte:   0,
			})
			_, diags = cfg.parser.decodeProvisioner(provBlock, nil)

			if !diags.HasErrors() {
				if !test.expectError {
					return
				}

				t.Fatalf("unexpected success")
			}

			if !test.expectError {
				for _, d := range diags {
					t.Errorf("%s", d)
				}
			}

			gotExpectedErr := false
			for _, d := range diags {
				if d.Summary == test.expectedErrorMessage {
					gotExpectedErr = true
				}

				t.Logf("got error (expected): '%s'", d.Summary)
			}

			if !gotExpectedErr {
				t.Errorf("never got expected error: '%s'", test.expectedErrorMessage)
			}
		})
	}
}
