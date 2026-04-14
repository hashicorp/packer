// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"testing"

	"github.com/hashicorp/packer/internal/enforcedparser"
)

func TestGetCoreBuildProvisionerFromBlock_AppliesOverrideForBuild(t *testing.T) {
	parser := getBasicParser()
	cfg := &PackerConfig{
		parser:                  parser,
		CorePackerVersionString: lockedVersion,
	}

	blocks, diags := enforcedparser.ParseProvisionerBlocks(`
provisioner "shell" {
  override = {
    "amazon-ebs.ubuntu" = {
      bool = false
    }
  }
}
`)
	if diags.HasErrors() {
		t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
	}

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	coreProv, diags := cfg.GetCoreBuildProvisionerFromEnforcedBlock(blocks[0], "amazon-ebs.ubuntu")
	if diags.HasErrors() {
		t.Fatalf("GetCoreBuildProvisionerFromBlock() unexpected error: %v", diags)
	}

	hclProv, ok := coreProv.Provisioner.(*HCL2Provisioner)
	if !ok {
		t.Fatalf("expected *HCL2Provisioner, got %T", coreProv.Provisioner)
	}

	if hclProv.override == nil {
		t.Fatal("expected override to be applied, got nil")
	}

	if got, ok := hclProv.override["bool"]; !ok || got != false {
		t.Fatalf("expected override bool=false, got %#v", hclProv.override["bool"])
	}
}

func TestGetCoreBuildProvisionerFromBlock_OverrideNotAppliedForOtherBuild(t *testing.T) {
	parser := getBasicParser()
	cfg := &PackerConfig{
		parser:                  parser,
		CorePackerVersionString: lockedVersion,
	}

	blocks, diags := enforcedparser.ParseProvisionerBlocks(`
provisioner "shell" {
  override = {
    "amazon-ebs.ubuntu" = {
      bool = false
    }
  }
}
`)
	if diags.HasErrors() {
		t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
	}

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	coreProv, diags := cfg.GetCoreBuildProvisionerFromEnforcedBlock(blocks[0], "virtualbox-iso.base")
	if diags.HasErrors() {
		t.Fatalf("GetCoreBuildProvisionerFromBlock() unexpected error: %v", diags)
	}

	hclProv, ok := coreProv.Provisioner.(*HCL2Provisioner)
	if !ok {
		t.Fatalf("expected *HCL2Provisioner, got %T", coreProv.Provisioner)
	}

	if hclProv.override != nil {
		t.Fatalf("expected no override to be applied, got %#v", hclProv.override)
	}
}

func TestGetCoreBuildProvisionerFromBlock_IncludesSensitiveVariables(t *testing.T) {
	parser := getBasicParser()
	cfg := &PackerConfig{
		parser:                  parser,
		CorePackerVersionString: lockedVersion,
		InputVariables: Variables{
			"visible": &Variable{Name: "visible"},
			"secret":  &Variable{Name: "secret", Sensitive: true},
		},
	}

	blocks, diags := enforcedparser.ParseProvisionerBlocks(`
provisioner "shell" {
	override = {
	  "amazon-ebs.ubuntu" = {
	    bool = false
	  }
	}
}
`)
	if diags.HasErrors() {
		t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
	}

	coreProv, diags := cfg.GetCoreBuildProvisionerFromEnforcedBlock(blocks[0], "amazon-ebs.ubuntu")
	if diags.HasErrors() {
		t.Fatalf("GetCoreBuildProvisionerFromBlock() unexpected error: %v", diags)
	}

	hclProv, ok := coreProv.Provisioner.(*HCL2Provisioner)
	if !ok {
		t.Fatalf("expected *HCL2Provisioner, got %T", coreProv.Provisioner)
	}

	sensitiveVars, ok := hclProv.builderVariables["packer_sensitive_variables"].([]string)
	if !ok {
		t.Fatalf("expected []string packer_sensitive_variables, got %T", hclProv.builderVariables["packer_sensitive_variables"])
	}

	if len(sensitiveVars) != 1 || sensitiveVars[0] != "secret" {
		t.Fatalf("expected sensitive vars [secret], got %#v", sensitiveVars)
	}
}

func TestParseProvisionerBlocks(t *testing.T) {
	tests := []struct {
		name         string
		blockContent string
		wantCount    int
		wantTypes    []string
		wantErr      bool
	}{
		{
			name: "single shell provisioner",
			blockContent: `
provisioner "shell" {
  inline = ["echo 'Hello from enforced provisioner'"]
}
`,
			wantCount: 1,
			wantTypes: []string{"shell"},
			wantErr:   false,
		},
		{
			name: "multiple provisioners",
			blockContent: `
provisioner "shell" {
  inline = ["echo 'First enforced provisioner'"]
}

provisioner "shell" {
  name   = "security-scan"
  inline = ["echo 'Security scan running...'"]
}
`,
			wantCount: 2,
			wantTypes: []string{"shell", "shell"},
			wantErr:   false,
		},
		{
			name: "provisioner with pause_before",
			blockContent: `
provisioner "shell" {
  pause_before = "10s"
  inline       = ["echo 'Waiting before execution'"]
}
`,
			wantCount: 1,
			wantTypes: []string{"shell"},
			wantErr:   false,
		},
		{
			name: "provisioner with max_retries",
			blockContent: `
provisioner "shell" {
  max_retries = 3
  inline      = ["echo 'Retry test'"]
}
`,
			wantCount: 1,
			wantTypes: []string{"shell"},
			wantErr:   false,
		},
		{
			name: "provisioner with only filter",
			blockContent: `
provisioner "shell" {
  only   = ["amazon-ebs.ubuntu"]
  inline = ["echo 'Only for amazon-ebs.ubuntu'"]
}
`,
			wantCount: 1,
			wantTypes: []string{"shell"},
			wantErr:   false,
		},
		{
			name: "provisioner with except filter",
			blockContent: `
provisioner "shell" {
  except = ["null.test"]
  inline = ["echo 'Except for null.test'"]
}
`,
			wantCount: 1,
			wantTypes: []string{"shell"},
			wantErr:   false,
		},
		{
			name:         "empty block content",
			blockContent: "",
			wantCount:    0,
			wantTypes:    nil,
			wantErr:      false,
		},
		{
			name:         "invalid HCL syntax",
			blockContent: "this is not valid { hcl }}}",
			wantCount:    0,
			wantTypes:    nil,
			wantErr:      true,
		},
		{
			name: "json single shell provisioner",
			blockContent: `{
  "provisioner": [
    {
      "shell": {
        "inline": ["echo 'Hello from enforced provisioner JSON'"]
      }
    }
  ]
}`,
			wantCount: 1,
			wantTypes: []string{"shell"},
			wantErr:   false,
		},
		{
			name: "json multiple provisioners",
			blockContent: `{
  "provisioner": [
    {
      "shell": {
        "inline": ["echo 'first'"]
      }
    },
    {
      "shell": {
        "name": "security-scan",
        "inline": ["echo 'second'"]
      }
    }
  ]
}`,
			wantCount: 2,
			wantTypes: []string{"shell", "shell"},
			wantErr:   false,
		},
		{
			name:         "invalid json syntax",
			blockContent: `{"provisioner": [ { "shell": { "inline": ["test"] } ] }`,
			wantCount:    0,
			wantTypes:    nil,
			wantErr:      true,
		},
		{
			name: "legacy json provisioners format",
			blockContent: `{
  "provisioners": [
    {
      "type": "shell",
      "inline": ["echo legacy json format"]
    }
  ]
}`,
			wantCount: 1,
			wantTypes: []string{"shell"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks, diags := enforcedparser.ParseProvisionerBlocks(tt.blockContent)

			if tt.wantErr {
				if !diags.HasErrors() {
					t.Errorf("ParseProvisionerBlocks() expected error but got none")
				}
				return
			}

			if diags.HasErrors() {
				t.Errorf("ParseProvisionerBlocks() unexpected error: %v", diags)
				return
			}

			if len(blocks) != tt.wantCount {
				t.Errorf("ParseProvisionerBlocks() got %d blocks, want %d", len(blocks), tt.wantCount)
				return
			}

			for i, wantType := range tt.wantTypes {
				if blocks[i].PType != wantType {
					t.Errorf("ParseProvisionerBlocks() block[%d].PType = %q, want %q", i, blocks[i].PType, wantType)
				}
			}
		})
	}
}

func TestParseProvisionerBlocksWithPauseBefore(t *testing.T) {
	blockContent := `
provisioner "shell" {
  pause_before = "30s"
  inline       = ["echo 'test'"]
}
`
	blocks, diags := enforcedparser.ParseProvisionerBlocks(blockContent)
	if diags.HasErrors() {
		t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	// pause_before should be parsed as 30 seconds
	if blocks[0].PauseBefore.Seconds() != 30 {
		t.Errorf("Expected PauseBefore=30s, got %v", blocks[0].PauseBefore)
	}
}

func TestParseProvisionerBlocksWithMaxRetries(t *testing.T) {
	blockContent := `
provisioner "shell" {
  max_retries = 5
  inline      = ["echo 'test'"]
}
`
	blocks, diags := enforcedparser.ParseProvisionerBlocks(blockContent)
	if diags.HasErrors() {
		t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	if blocks[0].MaxRetries != 5 {
		t.Errorf("Expected MaxRetries=5, got %d", blocks[0].MaxRetries)
	}
}

func TestParseProvisionerBlocksWithOnlyExcept(t *testing.T) {
	blockContent := `
provisioner "shell" {
  only   = ["amazon-ebs.ubuntu", "azure-arm.windows"]
  inline = ["echo 'test'"]
}
`
	blocks, diags := enforcedparser.ParseProvisionerBlocks(blockContent)
	if diags.HasErrors() {
		t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	// Check only filter
	if len(blocks[0].OnlyExcept.Only) != 2 {
		t.Errorf("Expected 2 only values, got %d", len(blocks[0].OnlyExcept.Only))
	}

	// Skip should return true for sources not in the only list
	if !blocks[0].OnlyExcept.Skip("null.test") {
		t.Error("Skip() should return true for source not in only list")
	}

	// Skip should return false for sources in the only list
	if blocks[0].OnlyExcept.Skip("amazon-ebs.ubuntu") {
		t.Error("Skip() should return false for source in only list")
	}
}

func TestParseProvisionerBlocksJSONWithOptions(t *testing.T) {
	blockContent := `{
  "provisioner": [
    {
      "shell": {
        "pause_before": "15s",
        "max_retries": 2,
        "only": ["docker.ubuntu"],
        "inline": ["echo 'json test'"]
      }
    }
  ]
}`

	blocks, diags := enforcedparser.ParseProvisionerBlocks(blockContent)
	if diags.HasErrors() {
		t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
	}

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	if blocks[0].PauseBefore.Seconds() != 15 {
		t.Errorf("Expected PauseBefore=15s, got %v", blocks[0].PauseBefore)
	}

	if blocks[0].MaxRetries != 2 {
		t.Errorf("Expected MaxRetries=2, got %d", blocks[0].MaxRetries)
	}

	if blocks[0].OnlyExcept.Skip("docker.ubuntu") {
		t.Error("Skip() should return false for source in only list")
	}

	if !blocks[0].OnlyExcept.Skip("null.test") {
		t.Error("Skip() should return true for source not in only list")
	}
}
