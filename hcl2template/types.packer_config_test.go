package hcl2template

import (
	"github.com/zclconf/go-cty/cty"
	"testing"

	"github.com/hashicorp/packer/packer"
)

var (
	refVBIsoUbuntu1204  = SourceRef{Type: "virtualbox-iso", Name: "ubuntu-1204"}
	refAWSEBSUbuntu1204 = SourceRef{Type: "amazon-ebs", Name: "ubuntu-1604"}
)

func TestParser_complete(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"working build",
			defaultParser,
			parseTestArgs{"testdata/complete", nil},
			&PackerConfig{
				Basedir: "testdata/complete",
				InputVariables: Variables{
					"foo": &Variable{
						DefaultValue: cty.StringVal("value"),
					},
					"image_id": &Variable{
						DefaultValue: cty.StringVal("image-id-default"),
					},
					"port": &Variable{
						DefaultValue: cty.NumberIntVal(42),
					},
					"availability_zone_names": &Variable{
						DefaultValue: cty.ListVal([]cty.Value{
							cty.StringVal("A"),
							cty.StringVal("B"),
							cty.StringVal("C"),
						}),
					},
				},
				LocalVariables: Variables{
					"feefoo": &Variable{
						DefaultValue: cty.StringVal("value_image-id-default"),
					},
					"standard_tags": &Variable{
						DefaultValue: cty.ObjectVal(map[string]cty.Value{
							"Component":   cty.StringVal("user-service"),
							"Environment": cty.StringVal("production"),
						}),
					},
					"abc_map": &Variable{
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
				Sources: map[SourceRef]*SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{refVBIsoUbuntu1204},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "shell",
								PName: "provisioner that does something",
							},
							{PType: "file"},
						},
						PostProcessors: []*PostProcessorBlock{
							{
								PType: "amazon-import",
								PName: "something",
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
					Type:     "virtualbox-iso",
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
								PType:         "amazon-import",
								PName:         "something",
								PostProcessor: basicMockPostProcessor,
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
		{"dir with no config files",
			defaultParser,
			parseTestArgs{"testdata/empty", nil},
			nil,
			true, true,
			nil,
			false,
		},
		{name: "inexistent dir",
			parser:                 defaultParser,
			args:                   parseTestArgs{"testdata/inexistent", nil},
			parseWantCfg:           nil,
			parseWantDiags:         true,
			parseWantDiagHasErrors: true,
		},
		{name: "folder named build.pkr.hcl with an unknown src",
			parser: defaultParser,
			args:   parseTestArgs{"testdata/build.pkr.hcl", nil},
			parseWantCfg: &PackerConfig{
				Basedir: "testdata/build.pkr.hcl",
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{refAWSEBSUbuntu1204, refVBIsoUbuntu1204},
						ProvisionerBlocks: []*ProvisionerBlock{
							{PType: "shell"},
							{PType: "file"},
						},
						PostProcessors: []*PostProcessorBlock{
							{PType: "amazon-import"},
						},
					},
				},
			},
			parseWantDiags:         false,
			parseWantDiagHasErrors: false,
			getBuildsWantBuilds:    []packer.Build{},
			getBuildsWantDiags:     true,
		},
		{name: "unknown block type",
			parser: defaultParser,
			args:   parseTestArgs{"testdata/unknown", nil},
			parseWantCfg: &PackerConfig{
				Basedir: "testdata/unknown",
			},
			parseWantDiags:         true,
			parseWantDiagHasErrors: true,
		},
	}
	testParse(t, tests)
}
