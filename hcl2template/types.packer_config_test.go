package hcl2template

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

var ref = SourceRef{Type: "virtualbox-iso", Name: "ubuntu-1204"}

func TestParser_complete(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"working build",
			defaultParser,
			parseTestArgs{"testdata/complete"},
			&PackerConfig{
				Sources: map[SourceRef]*Source{
					ref: &Source{Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Froms: []SourceRef{ref},
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
	}
	testParse(t, tests)
}
