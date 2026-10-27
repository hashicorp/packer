// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/packer/builder/null"
	"github.com/hashicorp/packer/packer"
)

func TestParse_source(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"two basic sources",
			defaultParser,
			parseTestArgs{"testdata/sources/basic.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
				Basedir: filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					},
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				{
					Type:           "null.test",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					SensitiveVars:  []string{},
				},
			},
			false,
			nil,
		},
		{"untyped source",
			defaultParser,
			parseTestArgs{"testdata/sources/untyped.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
			},
			true, true,
			nil,
			false,
			nil,
		},
		{"unnamed source",
			defaultParser,
			parseTestArgs{"testdata/sources/unnamed.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
			},
			true, true,
			nil,
			false,
			nil,
		},
		{"unused source with unknown type fails",
			defaultParser,
			parseTestArgs{"testdata/sources/nonexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Builds:                  nil,
				Basedir:                 filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{Type: "nonexistent", Name: "ubuntu-1204"}: {Type: "nonexistent", Name: "ubuntu-1204"},
				},
			},
			true, true,
			[]*packer.CoreBuild{},
			false,
			nil,
		},
		{"used source with unknown type fails",
			defaultParser,
			parseTestArgs{"testdata/sources/nonexistent_used.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{Type: "nonexistent", Name: "ubuntu-1204"}: {Type: "nonexistent", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{Type: "nonexistent", Name: "ubuntu-1204"},
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
			nil,
		},
		{"duplicate source",
			defaultParser,
			parseTestArgs{"testdata/sources/duplicate.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "sources"),
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					}: {
						Type: "virtualbox-iso",
						Name: "ubuntu-1204",
					},
				},
			},
			true, true,
			nil,
			false,
			nil,
		},
	}
	testParse(t, tests)
}

func TestSourceUseBlock_builderVariables(t *testing.T) {
	tests := []struct {
		name     string
		source   SourceUseBlock
		expected map[string]interface{}
	}{
		{
			name: "default source name",
			source: SourceUseBlock{
				SourceRef: SourceRef{
					Type: "amazon-ebs",
					Name: "source1",
				},
			},
			expected: map[string]interface{}{
				"packer_build_name":   "source1",
				"packer_builder_type": "amazon-ebs",
			},
		},
		{
			name: "overridden source name",
			source: SourceUseBlock{
				SourceRef: SourceRef{
					Type: "amazon-ebs",
					Name: "source1",
				},
				LocalName: "workshop",
			},
			expected: map[string]interface{}{
				"packer_build_name":   "workshop",
				"packer_builder_type": "amazon-ebs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.source.builderVariables()
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Fatalf("builderVariables() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPackerConfig_startProvisioner_buildName(t *testing.T) {
	tests := []struct {
		name         string
		source       SourceUseBlock
		expectedName string
	}{
		{
			name: "default source name",
			source: SourceUseBlock{
				SourceRef: SourceRef{
					Type: "null",
					Name: "source1",
				},
			},
			expectedName: "source1",
		},
		{
			name: "overridden source name",
			source: SourceUseBlock{
				SourceRef: SourceRef{
					Type: "null",
					Name: "source1",
				},
				LocalName: "workshop",
			},
			expectedName: "workshop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := getBasicParser()
			cfg := &PackerConfig{
				parser: parser,
			}
			pb := &ProvisionerBlock{
				PType:   "shell",
				HCL2Ref: HCL2Ref{Rest: &hclsyntax.Body{}},
			}
			ectx := cfg.EvalContext(BuildContext, nil)
			prov, diags := cfg.startProvisioner(tt.source, pb, ectx)
			if diags.HasErrors() {
				t.Fatalf("unexpected startProvisioner errors: %s", diags)
			}

			hclProv, ok := prov.(*HCL2Provisioner)
			if !ok {
				t.Fatalf("expected *HCL2Provisioner, got %T", prov)
			}

			gotName, ok := hclProv.builderVariables["packer_build_name"].(string)
			if !ok {
				t.Fatalf("expected string packer_build_name, got %T", hclProv.builderVariables["packer_build_name"])
			}

			if gotName != tt.expectedName {
				t.Fatalf("expected packer_build_name to be %q, got %q", tt.expectedName, gotName)
			}
		})
	}
}

func TestPackerConfig_startPostProcessor_buildName(t *testing.T) {
	tests := []struct {
		name         string
		source       SourceUseBlock
		expectedName string
	}{
		{
			name: "default source name",
			source: SourceUseBlock{
				SourceRef: SourceRef{
					Type: "null",
					Name: "source1",
				},
			},
			expectedName: "source1",
		},
		{
			name: "overridden source name",
			source: SourceUseBlock{
				SourceRef: SourceRef{
					Type: "null",
					Name: "source1",
				},
				LocalName: "workshop",
			},
			expectedName: "workshop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := getBasicParser()
			cfg := &PackerConfig{
				parser: parser,
			}
			pp := &PostProcessorBlock{
				PType:   "manifest",
				HCL2Ref: HCL2Ref{Rest: &hclsyntax.Body{}},
			}
			ectx := cfg.EvalContext(BuildContext, nil)
			postProc, diags := cfg.startPostProcessor(tt.source, pp, ectx)
			if diags.HasErrors() {
				t.Fatalf("unexpected startPostProcessor errors: %s", diags)
			}

			hclPP, ok := postProc.(*HCL2PostProcessor)
			if !ok {
				t.Fatalf("expected *HCL2PostProcessor, got %T", postProc)
			}

			gotName, ok := hclPP.builderVariables["packer_build_name"].(string)
			if !ok {
				t.Fatalf("expected string packer_build_name, got %T", hclPP.builderVariables["packer_build_name"])
			}

			if gotName != tt.expectedName {
				t.Fatalf("expected packer_build_name to be %q, got %q", tt.expectedName, gotName)
			}
		})
	}
}
