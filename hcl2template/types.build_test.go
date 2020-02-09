package hcl2template

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestParse_build(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic build no src",
			defaultParser,
			parseTestArgs{"testdata/build/basic.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{
							{
								Type: "amazon-ebs",
								Name: "ubuntu-1604",
							},
							refVBIsoUbuntu1204,
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
			parseTestArgs{"testdata/build/provisioner_untyped.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			nil,
			false,
		},
		{"inexistent provisioner",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_inexistent.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			[]packer.Build{&packer.CoreBuild{}},
			false,
		},
		{"untyped post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_untyped.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			[]packer.Build{&packer.CoreBuild{}},
			false,
		},
		{"inexistent post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_inexistent.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			[]packer.Build{},
			false,
		},
		{"invalid source",
			defaultParser,
			parseTestArgs{"testdata/build/invalid_source_reference.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			[]packer.Build{},
			false,
		},
	}
	testParse(t, tests)
}
