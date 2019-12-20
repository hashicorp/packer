package hcl2template

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestParse_build(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic build no src",
			defaultParser,
			parseTestArgs{"testdata/build/basic.pkr.hcl"},
			&PackerConfig{
				Builds: Builds{
					&BuildBlock{
						Froms: []SourceRef{
							{
								Type: "amazon-ebs",
								Name: "ubuntu-1604",
							},
							ref,
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "shell",
							},
							{
								PType: "file",
							},
						},
						PostProcessors: []*PostProcessorBlock{
							{
								PType: "amazon-import",
							},
						},
					},
				},
			},
			false, false,
			[]packer.Build{},
			true,
		},
		{"untyped provisioner",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_untyped.pkr.hcl"},
			&PackerConfig{
				Builds: nil,
			},
			true, true,
			nil,
			false,
		},
		{"inexistent provisioner",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_inexistent.pkr.hcl"},
			&PackerConfig{
				Builds: nil,
			},
			true, true,
			[]packer.Build{&packer.CoreBuild{}},
			false,
		},
		{"untyped post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_untyped.pkr.hcl"},
			&PackerConfig{
				Builds: nil,
			},
			true, true,
			[]packer.Build{&packer.CoreBuild{}},
			false,
		},
		{"inexistent post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_inexistent.pkr.hcl"},
			&PackerConfig{
				Builds: nil,
			},
			true, true,
			[]packer.Build{},
			false,
		},
	}
	testParse(t, tests)
}
