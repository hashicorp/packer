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
						Name:         "foo",
						DefaultValue: cty.StringVal("value"),
					},
					"image_id": &Variable{
						Name:         "image_id",
						DefaultValue: cty.StringVal("image-id-default"),
					},
					"port": &Variable{
						Name:         "port",
						DefaultValue: cty.NumberIntVal(42),
					},
					"availability_zone_names": &Variable{
						Name: "availability_zone_names",
						DefaultValue: cty.ListVal([]cty.Value{
							cty.StringVal("A"),
							cty.StringVal("B"),
							cty.StringVal("C"),
						}),
					},
				},
				LocalVariables: Variables{
					"feefoo": &Variable{
						Name:         "feefoo",
						DefaultValue: cty.StringVal("value_image-id-default"),
					},
					"standard_tags": &Variable{
						Name: "standard_tags",
						DefaultValue: cty.ObjectVal(map[string]cty.Value{
							"Component":   cty.StringVal("user-service"),
							"Environment": cty.StringVal("production"),
						}),
					},
					"abc_map": &Variable{
						Name: "abc_map",
						DefaultValue: cty.TupleVal([]cty.Value{
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
						PostProcessors: []*PostProcessorBlock{
							{
								PType:             "amazon-import",
								PName:             "something",
								KeepInputArtifact: pTrue,
							},
							{
								PType: "amazon-import",
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
							PType:       "shell",
							PName:       "provisioner that does something",
							Provisioner: basicMockProvisioner,
						},
						{PType: "file", Provisioner: basicMockProvisioner},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType:             "amazon-import",
								PName:             "something",
								PostProcessor:     basicMockPostProcessor,
								KeepInputArtifact: pTrue,
							},
							{
								PType:         "amazon-import",
								PostProcessor: basicMockPostProcessor,
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
							PType:       "shell",
							PName:       "provisioner that does something",
							Provisioner: basicMockProvisioner,
						},
						{PType: "file", Provisioner: basicMockProvisioner},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType:             "amazon-import",
								PName:             "something",
								PostProcessor:     basicMockPostProcessor,
								KeepInputArtifact: pTrue,
							},
							{
								PType:         "amazon-import",
								PostProcessor: basicMockPostProcessor,
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
