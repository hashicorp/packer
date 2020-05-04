package hcl2template

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
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
			parseTestArgs{"testdata/empty", nil, nil},
			nil,
			true, true,
			nil,
			false,
		},
		{name: "inexistent dir",
			parser:                 defaultParser,
			args:                   parseTestArgs{"testdata/inexistent", nil, nil},
			parseWantCfg:           nil,
			parseWantDiags:         true,
			parseWantDiagHasErrors: true,
		},
		{name: "folder named build.pkr.hcl with an unknown src",
			parser: defaultParser,
			args:   parseTestArgs{"testdata/build.pkr.hcl", nil, nil},
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
			args:   parseTestArgs{"testdata/unknown", nil, nil},
			parseWantCfg: &PackerConfig{
				Basedir: "testdata/unknown",
			},
			parseWantDiags:         true,
			parseWantDiagHasErrors: true,
		},
		{"provisioner with wrappers pause_before and max_retriers",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_paused_before_retry.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Sources: map[SourceRef]*SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{refVBIsoUbuntu1204},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType:       "shell",
								PauseBefore: time.Second * 10,
								MaxRetries:  5,
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
					Builder:  emptyMockBuilder,
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PType: "shell",
							Provisioner: &packer.RetriedProvisioner{
								MaxRetries: 5,
								Provisioner: &packer.PausedProvisioner{
									PauseBefore: time.Second * 10,
									Provisioner: emptyMockProvisioner,
								},
							},
						},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
				},
			},
			false,
		},
		{"provisioner with wrappers timeout",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_timeout.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Sources: map[SourceRef]*SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{refVBIsoUbuntu1204},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType:   "shell",
								Timeout: time.Second * 10,
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
					Builder:  emptyMockBuilder,
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PType: "shell",
							Provisioner: &packer.TimeoutProvisioner{
								Timeout:     time.Second * 10,
								Provisioner: emptyMockProvisioner,
							},
						},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
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
