package hcl2template

import (
	"testing"

	. "github.com/hashicorp/packer/hcl2template/internal"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

var (
	refVBIsoUbuntu1204  = SourceRef{Type: "virtualbox-iso", Name: "ubuntu-1204"}
	refAWSEBSUbuntu1604 = SourceRef{Type: "amazon-ebs", Name: "ubuntu-1604"}
	pTrue               = pointerToBool(true)
)

func TestParser_complete(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"working build",
			defaultParser,
			parseTestArgs{"testdata/complete", nil, nil},
			&PackerConfig{
				Basedir: "testdata/complete",
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
				},
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204:  {Type: "virtualbox-iso", Name: "ubuntu-1204"},
					refAWSEBSUbuntu1604: {Type: "amazon-ebs", Name: "ubuntu-1604"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{
							refVBIsoUbuntu1204,
							refAWSEBSUbuntu1604,
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "shell",
								PName: "provisioner that does something",
							},
							{PType: "file"},
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
			[]packer.Build{
				&packer.CoreBuild{
					Type:     "virtualbox-iso.ubuntu-1204",
					Prepared: true,
					Builder:  basicMockBuilder,
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
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType: "amazon-import",
								PName: "something",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
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
					Type:     "amazon-ebs.ubuntu-1604",
					Prepared: true,
					Builder: &MockBuilder{
						Config: MockConfig{
							NestedMockConfig: NestedMockConfig{
								String: "setting from build section",
								Int:    42,
								Tags:   []MockTag{},
							},
							NestedSlice: []NestedMockConfig{},
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
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType: "amazon-import",
								PName: "something",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: basicMockPostProcessor,
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

func pointerToBool(b bool) *bool {
	return &b
}
