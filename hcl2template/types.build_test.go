package hcl2template

import (
	"path/filepath"
	"testing"

	. "github.com/hashicorp/packer/hcl2template/internal"
	"github.com/hashicorp/packer/packer"
)

func TestParse_build(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic build no src",
			defaultParser,
			parseTestArgs{"testdata/build/basic.pkr.hcl", nil, nil},
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
			parseTestArgs{"testdata/build/provisioner_untyped.pkr.hcl", nil, nil},
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
			parseTestArgs{"testdata/build/provisioner_inexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			[]packer.Build{&CoreHCL2Build{}},
			false,
		},
		{"untyped post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_untyped.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			[]packer.Build{&CoreHCL2Build{}},
			false,
		},
		{"inexistent post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_inexistent.pkr.hcl", nil, nil},
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
			parseTestArgs{"testdata/build/invalid_source_reference.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds:  nil,
			},
			true, true,
			[]packer.Build{},
			false,
		},
		{"named build",
			defaultParser,
			parseTestArgs{"testdata/build/named.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Builds: Builds{
					&BuildBlock{
						Name: "somebuild",
						Sources: []SourceRef{
							{
								Type: "amazon-ebs",
								Name: "ubuntu-1604",
							},
							refVBIsoUbuntu1204,
						},
					},
				},
			},
			false, false,
			[]packer.Build{},
			true,
		},
		{"post-processor with only and except",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_onlyexcept.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204:  {Type: "virtualbox-iso", Name: "ubuntu-1204"},
					refAWSEBSUbuntu1604: {Type: "amazon-ebs", Name: "ubuntu-1604"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources:           []SourceRef{refVBIsoUbuntu1204, refAWSEBSUbuntu1604},
						ProvisionerBlocks: nil,
						PostProcessors: []*PostProcessorBlock{
							{
								PType:      "amazon-import",
								OnlyExcept: OnlyExcept{Only: []string{"virtualbox-iso.ubuntu-1204"}, Except: nil},
							},
							{
								PType:      "manifest",
								OnlyExcept: OnlyExcept{Only: nil, Except: []string{"virtualbox-iso.ubuntu-1204"}},
							},
						},
					},
				},
			},
			false, false,
			[]packer.Build{
				&CoreHCL2Build{
					Type:         "virtualbox-iso.ubuntu-1204",
					Builder:      emptyMockBuilder,
					Provisioners: []CoreHCL2BuildProvisioner{},
					PostProcessors: [][]CoreHCL2BuildPostProcessor{
						{
							{
								PType: "amazon-import",
								PostProcessor: &MockPostProcessor{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
										NestedSlice:      []NestedMockConfig{},
									},
								},
							},
						},
					},
				},
				&CoreHCL2Build{
					Type:         "amazon-ebs.ubuntu-1604",
					Builder:      emptyMockBuilder,
					Provisioners: []CoreHCL2BuildProvisioner{},
					PostProcessors: [][]CoreHCL2BuildPostProcessor{
						{
							{
								PType: "manifest",
								PostProcessor: &MockPostProcessor{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
										NestedSlice:      []NestedMockConfig{},
									},
								},
							},
						},
					},
				},
			},
			false,
		},
		{"provisioner with only and except",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_onlyexcept.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204:  {Type: "virtualbox-iso", Name: "ubuntu-1204"},
					refAWSEBSUbuntu1604: {Type: "amazon-ebs", Name: "ubuntu-1604"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{refVBIsoUbuntu1204, refAWSEBSUbuntu1604},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType:      "shell",
								OnlyExcept: OnlyExcept{Only: []string{"virtualbox-iso.ubuntu-1204"}, Except: nil},
							},
							{
								PType:      "file",
								OnlyExcept: OnlyExcept{Only: nil, Except: []string{"virtualbox-iso.ubuntu-1204"}},
							},
						},
					},
				},
			},
			false, false,
			[]packer.Build{
				&CoreHCL2Build{
					Type:    "virtualbox-iso.ubuntu-1204",
					Builder: emptyMockBuilder,
					Provisioners: []CoreHCL2BuildProvisioner{
						{
							PType: "shell",
							Provisioner: &MockProvisioner{
								Config: MockConfig{
									NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
									NestedSlice:      []NestedMockConfig{},
								},
							},
						},
					},
					PostProcessors: [][]CoreHCL2BuildPostProcessor{},
				},
				&CoreHCL2Build{
					Type:    "amazon-ebs.ubuntu-1604",
					Builder: emptyMockBuilder,
					Provisioners: []CoreHCL2BuildProvisioner{
						{
							PType: "file",
							Provisioner: &MockProvisioner{
								Config: MockConfig{
									NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
									NestedSlice:      []NestedMockConfig{},
								},
							},
						},
					},
					PostProcessors: [][]CoreHCL2BuildPostProcessor{},
				},
			},
			false,
		},
	}
	testParse(t, tests)
}
