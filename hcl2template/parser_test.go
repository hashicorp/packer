// Copyright IBM Corp. 2026
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

func TestDetectBuildPrereqDependencies_wholeLocal(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       []refString
	}{
		{"whole object", "merge(var, local)", []refString{{MType: "local", Name: "version"}, {MType: "local", Name: "name"}}},
		{"explicit reference", "local.version", []refString{{MType: "local", Name: "version"}}},
		{"scoped local", "[for local in [1, 2] : local]", nil},
		{"scoped local attribute", "[for local in [{version = 1}] : local.version]", nil},
		{"explicit self reference", "local.result", []refString{{MType: "local", Name: "result"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, diags := hclsyntax.ParseExpression([]byte(tt.expression), "test.pkr.hcl", hcl.InitialPos)
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			result := &LocalBlock{LocalName: "result", Expr: expr}
			cfg := &PackerConfig{LocalBlocks: []*LocalBlock{
				{LocalName: "version", Expr: hcl.StaticExpr(cty.StringVal("14"), hcl.Range{})},
				{LocalName: "name", Expr: hcl.StaticExpr(cty.StringVal("image"), hcl.Range{})},
				result,
			}}
			if diags := cfg.detectBuildPrereqDependencies(); diags.HasErrors() {
				t.Fatal(diags)
			}
			if diff := cmp.Diff(tt.want, result.dependencies); diff != "" {
				t.Fatalf("unexpected dependencies (-want +got): %s", diff)
			}
		})
	}
}

func TestParse_wholeLocalDependencies(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		localName  string
		want       cty.Value
		wantCycle  bool
		datasource bool
	}{
		{
			name: "templatefile with datasource backed locals",
			template: `
variable "version" { default = "14" }
data "null" "macos" { input = var.version }
locals {
  version = { major = data.null.macos.output }
  vm_name = "image-${local.version.major}"
  rendered = templatefile("version.pkrtpl", merge(var, local))
}
`,
			localName: "rendered",
			want:      cty.StringVal("image-14:14"),
		},
		{
			name: "datasource with whole local input",
			template: `
variable "version" { default = "14" }
locals {
  version = { major = var.version }
  vm_name = "image-${local.version.major}"
}
data "null" "rendered" {
  input = templatefile("version.pkrtpl", merge(var, local))
}
`,
			datasource: true,
			want:       cty.StringVal("image-14:14"),
		},
		{
			name:      "empty local object",
			template:  `locals { result = length(local) }`,
			localName: "result",
			want:      cty.NumberIntVal(0),
		},
		{
			name: "local cycle",
			template: `
locals {
  result = merge(local)
  other = local.result
}
`,
			wantCycle: true,
		},
		{
			name: "datasource cycle",
			template: `
locals { result = data.null.rendered.output }
data "null" "rendered" { input = jsonencode(local) }
`,
			wantCycle: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			filename := filepath.Join(dir, "test.pkr.hcl")
			if err := os.WriteFile(filename, []byte(tt.template), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "version.pkrtpl"), []byte("${vm_name}:${version.major}"), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, diags := getBasicParser().Parse(filename, nil, nil)
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			diags = cfg.Initialize(packer.InitializeOptions{})
			if tt.wantCycle {
				if !diags.HasErrors() || !strings.Contains(diags.Error(), "Cycle:") {
					t.Fatalf("expected a cycle diagnostic, got %s", diags)
				}
				return
			}
			if diags.HasErrors() {
				t.Fatal(diags)
			}
			var got cty.Value
			if tt.datasource {
				ds := cfg.Datasources[DatasourceRef{Type: "null", Name: "rendered"}]
				wantDeps := []refString{{MType: "local", Name: "version"}, {MType: "local", Name: "vm_name"}}
				if diff := cmp.Diff(wantDeps, ds.Dependencies, cmpopts.SortSlices(func(a, b refString) bool { return a.String() < b.String() })); diff != "" {
					t.Fatalf("unexpected datasource dependencies (-want +got): %s", diff)
				}
				got = ds.value.GetAttr("output")
			} else {
				got = cfg.LocalVariables[tt.localName].Value()
			}
			if !got.RawEquals(tt.want) {
				t.Fatalf("got %s, want %s", got.GoString(), tt.want.GoString())
			}
		})
	}
}
