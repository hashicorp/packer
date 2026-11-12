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
		{"nested lookup", `upper(lookup(local, "version", "unknown"))`, []refString{{MType: "local", Name: "version"}}},
		{"lookup with whole local default", `lookup(local, "missing", merge(local))`, []refString{{MType: "local", Name: "version"}, {MType: "local", Name: "name"}}},
		{"dynamic lookup", `lookup(local, var.key, "unknown")`, []refString{{MType: "local", Name: "version"}, {MType: "local", Name: "name"}}},
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
		json       bool
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
			name: "independent whole local lookups",
			template: `
data "null" "independent" { input = lookup(local, "missing", "c") }
locals {
  first = lookup(local, "missing", "a")
  second = lookup(local, "missing", "b")
  result = [local.first, local.second, data.null.independent.output]
}
`,
			localName: "result",
			want:      cty.TupleVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b"), cty.StringVal("c")}),
		},
		{
			name: "JSON local name containing a dot",
			template: `{
  "locals": {
    "foo.bar": "a",
    "result": "${length(local)}"
  }
}`,
			localName: "result",
			want:      cty.NumberIntVal(1),
			json:      true,
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
			ext := hcl2FileExt
			if tt.json {
				ext = hcl2JsonFileExt
			}
			filename := filepath.Join(dir, "test"+ext)
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
