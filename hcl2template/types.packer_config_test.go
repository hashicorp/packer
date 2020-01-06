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
			parseTestArgs{"testdata/complete"},
			&PackerConfig{
				Sources: map[SourceRef]*Source{
					refVBIsoUbuntu1204: &Source{Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Froms: []SourceRef{refVBIsoUbuntu1204},
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
			false, false,
			[]packer.Build{
				&packer.CoreBuild{
					Type:    "virtualbox-iso",
					Builder: basicMockBuilder,
					Provisioners: []packer.CoreBuildProvisioner{
						{PType: "shell", Provisioner: basicMockProvisioner},
						{PType: "file", Provisioner: basicMockProvisioner},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{PType: "amazon-import", PostProcessor: basicMockPostProcessor},
						},
					},
				},
			},
			false,
		},
		{"dir with no config files",
			defaultParser,
			parseTestArgs{"testdata/empty"},
			nil,
			true, true,
			nil,
			false,
		},
		{name: "inexistent dir",
			parser:                 defaultParser,
			args:                   parseTestArgs{"testdata/inexistent"},
			parseWantCfg:           nil,
			parseWantDiags:         true,
			parseWantDiagHasErrors: true,
		},
		{name: "folder named build.pkr.hcl with an unknown src",
			parser: defaultParser,
			args:   parseTestArgs{"testdata/build.pkr.hcl"},
			parseWantCfg: &PackerConfig{
				Builds: Builds{
					&BuildBlock{
						Froms: []SourceRef{refAWSEBSUbuntu1204, refVBIsoUbuntu1204},
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
	}
	testParse(t, tests)
}
