package hcl2template

import (
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
					"foo":                     &Variable{},
					"image_id":                &Variable{},
					"port":                    &Variable{},
					"availability_zone_names": &Variable{},
				},
				LocalVariables: Variables{
					"feefoo": &Variable{},
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
