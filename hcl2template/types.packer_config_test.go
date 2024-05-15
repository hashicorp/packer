// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/hcl2template/addrs"
	. "github.com/hashicorp/packer/hcl2template/internal"
	hcl2template "github.com/hashicorp/packer/hcl2template/internal"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

var (
	refVBIsoUbuntu1204  = SourceRef{Type: "virtualbox-iso", Name: "ubuntu-1204"}
	refAWSEBSUbuntu1604 = SourceRef{Type: "amazon-ebs", Name: "ubuntu-1604"}
	refAWSV3MyImage     = SourceRef{Type: "amazon-v3-ebs", Name: "my-image"}
	refNull             = SourceRef{Type: "null", Name: "test"}
	pTrue               = pointerToBool(true)
)

func TestParser_complete(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"working build",
			defaultParser,
			parseTestArgs{"testdata/complete", nil, nil},
			&PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					VersionConstraints: []VersionConstraint{
						{
							Required: mustVersionConstraints(version.NewConstraint(">= v1")),
						},
					},
					RequiredPlugins: nil,
				},
				CorePackerVersionString: lockedVersion,
				Basedir:                 "testdata/complete",

				InputVariables: Variables{
					"foo": &Variable{
						Name:   "foo",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("value")}},
						Type:   cty.String,
					},
					"image_id": &Variable{
						Name:   "image_id",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("image-id-default")}},
						Type:   cty.String,
					},
					"port": &Variable{
						Name:   "port",
						Values: []VariableAssignment{{From: "default", Value: cty.NumberIntVal(42)}},
						Type:   cty.Number,
					},
					"availability_zone_names": &Variable{
						Name: "availability_zone_names",
						Values: []VariableAssignment{{
							From: "default",
							Value: cty.ListVal([]cty.Value{
								cty.StringVal("A"),
								cty.StringVal("B"),
								cty.StringVal("C"),
							}),
						}},
						Type: cty.List(cty.String),
					},
				},
				LocalVariables: Variables{
					"feefoo": &Variable{
						Name:   "feefoo",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("value_image-id-default")}},
						Type:   cty.String,
					},
					"data_source": &Variable{
						Name:   "data_source",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("string")}},
						Type:   cty.String,
					},
					"standard_tags": &Variable{
						Name: "standard_tags",
						Values: []VariableAssignment{{From: "default",
							Value: cty.ObjectVal(map[string]cty.Value{
								"Component":   cty.StringVal("user-service"),
								"Environment": cty.StringVal("production"),
							}),
						}},
						Type: cty.Object(map[string]cty.Type{
							"Component":   cty.String,
							"Environment": cty.String,
						}),
					},
					"abc_map": &Variable{
						Name: "abc_map",
						Values: []VariableAssignment{{From: "default",
							Value: cty.TupleVal([]cty.Value{
								cty.ObjectVal(map[string]cty.Value{
									"id": cty.StringVal("a"),
								}),
								cty.ObjectVal(map[string]cty.Value{
									"id": cty.StringVal("b"),
								}),
								cty.ObjectVal(map[string]cty.Value{
									"id": cty.StringVal("c"),
								}),
							}),
						}},
						Type: cty.Tuple([]cty.Type{
							cty.Object(map[string]cty.Type{
								"id": cty.String,
							}),
							cty.Object(map[string]cty.Type{
								"id": cty.String,
							}),
							cty.Object(map[string]cty.Type{
								"id": cty.String,
							}),
						}),
					},
					"supersecret": &Variable{
						Name: "supersecret",
						Values: []VariableAssignment{{From: "default",
							Value: cty.StringVal("image-id-default-password")}},
						Type:      cty.String,
						Sensitive: true,
					},
				},
				Datasources: Datasources{
					DatasourceRef{Type: "amazon-ami", Name: "test"}: DatasourceBlock{
						Type:  "amazon-ami",
						Name:  "test",
						value: cty.StringVal("foo"),
					},
				},
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204:  {Type: "virtualbox-iso", Name: "ubuntu-1204"},
					refAWSEBSUbuntu1604: {Type: "amazon-ebs", Name: "ubuntu-1604"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
							{
								SourceRef: refAWSEBSUbuntu1604,
							},
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "shell",
								PName: "provisioner that does something",
							},
							{PType: "file"},
						},
						ErrorCleanupProvisionerBlock: &ProvisionerBlock{
							PType: "shell",
							PName: "error-cleanup-provisioner that does something",
						},
						PostProcessorsLists: [][]*PostProcessorBlock{
							{
								{
									PType:             "amazon-import",
									PName:             "something",
									KeepInputArtifact: pTrue,
								},
							},
							{
								{
									PType: "amazon-import",
								},
							},
							{
								{
									PType: "amazon-import",
									PName: "first-nested-post-processor",
								},
								{
									PType: "amazon-import",
									PName: "second-nested-post-processor",
								},
							},
							{
								{
									PType: "amazon-import",
									PName: "third-nested-post-processor",
								},
								{
									PType: "amazon-import",
									PName: "fourth-nested-post-processor",
								},
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:        "virtualbox-iso.ubuntu-1204",
					BuilderType: "virtualbox-iso",
					Prepared:    true,
					Builder: &MockBuilder{
						Config: MockConfig{
							NestedMockConfig: NestedMockConfig{
								// interpolates source and type in builder
								String:   "ubuntu-1204-virtualbox-iso",
								Int:      42,
								Int64:    43,
								Bool:     true,
								Trilean:  config.TriTrue,
								Duration: 10 * time.Second,
								MapStringString: map[string]string{
									"a": "b",
									"c": "d",
								},
								SliceString: []string{
									"a",
									"b",
									"c",
								},
								SliceSliceString: [][]string{
									{"a", "b"},
									{"c", "d"},
								},
								Tags:       []MockTag{},
								Datasource: "string",
							},
							Nested: builderBasicNestedMockConfig,
							NestedSlice: []NestedMockConfig{
								builderBasicNestedMockConfig,
								builderBasicNestedMockConfig,
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PType: "shell",
							PName: "provisioner that does something",
							Provisioner: &HCL2Provisioner{
								Provisioner: basicMockProvisioner,
							},
						},
						{
							PType: "file",
							Provisioner: &HCL2Provisioner{
								Provisioner: basicMockProvisioner,
							},
						},
					},
					CleanupProvisioner: packer.CoreBuildProvisioner{
						PType: "shell",
						PName: "error-cleanup-provisioner that does something",
						Provisioner: &HCL2Provisioner{
							Provisioner: basicMockProvisioner,
						},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType: "amazon-import",
								PName: "something",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessorDynamicTags,
								},
								KeepInputArtifact: pTrue,
							},
						},
						{
							{
								PType: "amazon-import",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
						},
						{
							{
								PType: "amazon-import",
								PName: "first-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
							{
								PType: "amazon-import",
								PName: "second-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
						},
						{
							{
								PType: "amazon-import",
								PName: "third-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
							{
								PType: "amazon-import",
								PName: "fourth-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
						},
					},
				},
				&packer.CoreBuild{
					Type:        "amazon-ebs.ubuntu-1604",
					BuilderType: "amazon-ebs",
					Prepared:    true,
					Builder: &MockBuilder{
						Config: MockConfig{
							NestedMockConfig: NestedMockConfig{
								String: "setting from build section",
								Int:    42,
								Tags:   []MockTag{},
							},
							Nested: hcl2template.NestedMockConfig{
								Tags: []hcl2template.MockTag{
									{Key: "Component", Value: "user-service"},
									{Key: "Environment", Value: "production"},
								},
							},
							NestedSlice: []NestedMockConfig{
								hcl2template.NestedMockConfig{
									Tags: []hcl2template.MockTag{
										{Key: "Component", Value: "user-service"},
										{Key: "Environment", Value: "production"},
									},
								},
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PType: "shell",
							PName: "provisioner that does something",
							Provisioner: &HCL2Provisioner{
								Provisioner: basicMockProvisioner,
							},
						},
						{
							PType: "file",
							Provisioner: &HCL2Provisioner{
								Provisioner: basicMockProvisioner,
							},
						},
					},
					CleanupProvisioner: packer.CoreBuildProvisioner{
						PType: "shell",
						PName: "error-cleanup-provisioner that does something",
						Provisioner: &HCL2Provisioner{
							Provisioner: basicMockProvisioner,
						},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType: "amazon-import",
								PName: "something",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessorDynamicTags,
								},
								KeepInputArtifact: pTrue,
							},
						},
						{
							{
								PType: "amazon-import",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
						},
						{
							{
								PType: "amazon-import",
								PName: "first-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
							{
								PType: "amazon-import",
								PName: "second-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
						},
						{
							{
								PType: "amazon-import",
								PName: "third-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
							{
								PType: "amazon-import",
								PName: "fourth-nested-post-processor",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
								},
							},
						},
					},
				},
			},
			false,
		},
	}
	testParse(t, tests)
}

func TestParser_ValidateFilterOption(t *testing.T) {
	tests := []struct {
		pattern     string
		expectError bool
	}{
		{"*foo*", false},
		{"foo[]bar", true},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			_, diags := convertFilterOption([]string{test.pattern}, "")
			if diags.HasErrors() && !test.expectError {
				t.Fatalf("Expected %s to parse as glob", test.pattern)
			}
			if !diags.HasErrors() && test.expectError {
				t.Fatalf("Expected %s to fail to parse as glob", test.pattern)
			}
		})
	}
}

func TestParser_no_init(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"working build with imports",
			defaultParser,
			parseTestArgs{"testdata/init/imports", nil, nil},
			&PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					VersionConstraints: []VersionConstraint{
						{
							Required: mustVersionConstraints(version.NewConstraint(">= v1")),
						},
					},
					RequiredPlugins: []*RequiredPlugins{
						{
							RequiredPlugins: map[string]*RequiredPlugin{
								"amazon": {
									Name:   "amazon",
									Source: "github.com/hashicorp/amazon",
									Type: &addrs.Plugin{
										Source: "github.com/hashicorp/amazon",
									},
									Requirement: VersionConstraint{
										Required: mustVersionConstraints(version.NewConstraint(">= v0")),
									},
								},
								"amazon-v1": {
									Name:   "amazon-v1",
									Source: "github.com/hashicorp/amazon",
									Type: &addrs.Plugin{
										Source: "github.com/hashicorp/amazon",
									},
									Requirement: VersionConstraint{
										Required: mustVersionConstraints(version.NewConstraint(">= v1")),
									},
								},
								"amazon-v2": {
									Name:   "amazon-v2",
									Source: "github.com/hashicorp/amazon",
									Type: &addrs.Plugin{
										Source: "github.com/hashicorp/amazon",
									},
									Requirement: VersionConstraint{
										Required: mustVersionConstraints(version.NewConstraint(">= v2")),
									},
								},
								"amazon-v3": {
									Name:   "amazon-v3",
									Source: "github.com/hashicorp/amazon",
									Type: &addrs.Plugin{
										Source: "github.com/hashicorp/amazon",
									},
									Requirement: VersionConstraint{
										Required: mustVersionConstraints(version.NewConstraint(">= v3")),
									},
								},
								"amazon-v3-azr": {
									Name:   "amazon-v3-azr",
									Source: "github.com/azr/amazon",
									Type: &addrs.Plugin{
										Source: "github.com/azr/amazon",
									},
									Requirement: VersionConstraint{
										Required: mustVersionConstraints(version.NewConstraint(">= v3")),
									},
								},
								"amazon-v4": {
									Name:   "amazon-v4",
									Source: "github.com/hashicorp/amazon",
									Type: &addrs.Plugin{
										Source: "github.com/hashicorp/amazon",
									},
									Requirement: VersionConstraint{
										Required: mustVersionConstraints(version.NewConstraint(">= v4")),
									},
								},
							},
						},
					},
				},
				CorePackerVersionString: lockedVersion,
				Basedir:                 "testdata/init/imports",

				InputVariables: Variables{
					"foo": &Variable{
						Name:   "foo",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("value")}},
						Type:   cty.String,
					},
					"image_id": &Variable{
						Name:   "image_id",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("image-id-default")}},
						Type:   cty.String,
					},
					"port": &Variable{
						Name:   "port",
						Values: []VariableAssignment{{From: "default", Value: cty.NumberIntVal(42)}},
						Type:   cty.Number,
					},
					"availability_zone_names": &Variable{
						Name: "availability_zone_names",
						Values: []VariableAssignment{{
							From: "default",
							Value: cty.ListVal([]cty.Value{
								cty.StringVal("A"),
								cty.StringVal("B"),
								cty.StringVal("C"),
							}),
						}},
						Type: cty.List(cty.String),
					},
				},
				Sources: nil,
				Builds:  nil,
			},
			false, false,
			[]packersdk.Build{},
			false,
		},

		{"duplicate required plugin accessor fails",
			defaultParser,
			parseTestArgs{"testdata/init/duplicate_required_plugins", nil, nil},
			nil,
			true, true,
			[]packersdk.Build{},
			false,
		},
		{"invalid_inexplicit_source.pkr.hcl",
			defaultParser,
			parseTestArgs{"testdata/init/invalid_inexplicit_source.pkr.hcl", nil, nil},
			&PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					VersionConstraints: nil,
					RequiredPlugins: []*RequiredPlugins{
						{},
					},
				},
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Clean("testdata/init"),
			},
			true, true,
			[]packersdk.Build{},
			false,
		},
		{"invalid_short_source.pkr.hcl",
			defaultParser,
			parseTestArgs{"testdata/init/invalid_short_source.pkr.hcl", nil, nil},
			&PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					VersionConstraints: nil,
					RequiredPlugins: []*RequiredPlugins{
						{},
					},
				},
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Clean("testdata/init"),
			},
			true, true,
			[]packersdk.Build{},
			false,
		},
		{"invalid_inexplicit_source_2.pkr.hcl",
			defaultParser,
			parseTestArgs{"testdata/init/invalid_inexplicit_source_2.pkr.hcl", nil, nil},
			&PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					VersionConstraints: nil,
					RequiredPlugins: []*RequiredPlugins{
						{},
					},
				},
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Clean("testdata/init"),
			},
			true, true,
			[]packersdk.Build{},
			false,
		},
	}
	testParse_only_Parse(t, tests)
}

func pointerToBool(b bool) *bool {
	return &b
}

func mustVersionConstraints(vs version.Constraints, err error) version.Constraints {
	if err != nil {
		panic(err)
	}
	return vs
}
