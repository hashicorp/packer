// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package enforcedparser

import "testing"

func TestParseProvisionerBlocks_BasicFormats(t *testing.T) {
	tests := []struct {
		name         string
		blockContent string
		wantCount    int
		wantType     string
	}{
		{
			name: "hcl",
			blockContent: `
provisioner "shell" {
  inline = ["echo hello"]
}
`,
			wantCount: 1,
			wantType:  "shell",
		},
		{
			name: "hcl json",
			blockContent: `{
  "provisioner": [
    {
      "shell": {
        "inline": ["echo hello"]
      }
    }
  ]
}`,
			wantCount: 1,
			wantType:  "shell",
		},
		{
			name: "legacy json fallback",
			blockContent: `{
  "provisioners": [
    {
      "type": "shell",
      "inline": ["echo hello"]
    }
  ]
}`,
			wantCount: 1,
			wantType:  "shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks, diags := ParseProvisionerBlocks(tt.blockContent)
			if diags.HasErrors() {
				t.Fatalf("ParseProvisionerBlocks() unexpected error: %v", diags)
			}
			if len(blocks) != tt.wantCount {
				t.Fatalf("ParseProvisionerBlocks() got %d blocks, want %d", len(blocks), tt.wantCount)
			}
			if blocks[0].PType != tt.wantType {
				t.Fatalf("first block type = %q, want %q", blocks[0].PType, tt.wantType)
			}
		})
	}
}

func TestParseProvisionerBlocks_OverrideAndOnlyExcept(t *testing.T) {
	blocks, diags := ParseProvisionerBlocks(`
provisioner "shell" {
  only = ["amazon-ebs.ubuntu"]
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

	pb := blocks[0]
	if pb.OnlyExcept.Skip("amazon-ebs.ubuntu") {
		t.Fatal("Skip() should return false for source in only list")
	}
	if !pb.OnlyExcept.Skip("null.test") {
		t.Fatal("Skip() should return true for source not in only list")
	}

	rawOverride, ok := pb.Override["amazon-ebs.ubuntu"]
	if !ok {
		t.Fatal("expected override for amazon-ebs.ubuntu")
	}
	override, ok := rawOverride.(map[string]interface{})
	if !ok {
		t.Fatalf("override type = %T, want map[string]interface{}", rawOverride)
	}
	if got, ok := override["bool"]; !ok || got != false {
		t.Fatalf("override bool = %#v, want false", override["bool"])
	}
}

func TestParseProvisionerBlocks_InvalidContent(t *testing.T) {
	_, diags := ParseProvisionerBlocks("this is not valid { hcl }}}")
	if !diags.HasErrors() {
		t.Fatal("expected parse error, got none")
	}
}
