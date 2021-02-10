package hcl2template

import (
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
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
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "amazon-ebs",
									Name: "ubuntu-1604",
								},
							},
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "shell",
							},
							{
								PType: "file",
							},
						},
						PostProcessorsLists: [][]*PostProcessorBlock{
							{
								{
									PType: "amazon-import",
								},
							},
						},
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			true,
		},
		{"untyped provisioner",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_untyped.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Builds:                  nil,
			},
			true, true,
			nil,
			false,
		},
		{"inexistent provisioner",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_inexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Builds: Builds{
					&BuildBlock{
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "inexistant",
							},
						},
					},
				},
			},
			true, true,
			[]packersdk.Build{&packer.CoreBuild{
				Provisioners: []packer.CoreBuildProvisioner{},
			}},
			false,
		},
		{"two error-cleanup-provisioner",
			defaultParser,
			parseTestArgs{"testdata/build/two-error-cleanup-provisioner.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
			},
			true, true,
			[]packersdk.Build{&packer.CoreBuild{
				Builder: emptyMockBuilder,
				CleanupProvisioner: packer.CoreBuildProvisioner{
					PType: "shell-local",
					Provisioner: &HCL2Provisioner{
						Provisioner: &MockProvisioner{
							Config: MockConfig{
								NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
								NestedSlice:      []NestedMockConfig{},
							},
						},
					},
				},
			}},
			false,
		},
		{"untyped post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_untyped.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Builds:                  nil,
			},
			true, true,
			[]packersdk.Build{&packer.CoreBuild{}},
			false,
		},
		{"inexistent post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_inexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Builds: Builds{
					&BuildBlock{
						PostProcessorsLists: [][]*PostProcessorBlock{
							{
								{
									PType: "inexistant",
								},
							},
						},
					},
				},
			},
			true, true,
			[]packersdk.Build{&packer.CoreBuild{
				PostProcessors: [][]packer.CoreBuildPostProcessor{},
			}},
			true,
		},
		{"invalid source",
			defaultParser,
			parseTestArgs{"testdata/build/invalid_source_reference.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Builds:                  nil,
			},
			true, true,
			[]packersdk.Build{},
			false,
		},
		{"named build",
			defaultParser,
			parseTestArgs{"testdata/build/named.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Builds: Builds{
					&BuildBlock{
						Name: "somebuild",
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "amazon-ebs",
									Name: "ubuntu-1604",
								},
							},
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			true,
		},
		{"post-processor with only and except",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_onlyexcept.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204:  {Type: "virtualbox-iso", Name: "ubuntu-1204"},
					refAWSEBSUbuntu1604: {Type: "amazon-ebs", Name: "ubuntu-1604"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
							{
								SourceRef: SourceRef{Type: "amazon-ebs", Name: "ubuntu-1604"},
								LocalName: "aws-ubuntu-16.04",
							},
						},
						ProvisionerBlocks: nil,
						PostProcessorsLists: [][]*PostProcessorBlock{
							{
								{
									PType:      "amazon-import",
									OnlyExcept: OnlyExcept{Only: []string{"virtualbox-iso.ubuntu-1204"}, Except: nil},
								},
							},
							{
								{
									PType:      "manifest",
									OnlyExcept: OnlyExcept{Only: nil, Except: []string{"virtualbox-iso.ubuntu-1204"}},
								},
							},
							{
								{
									PType:      "amazon-import",
									OnlyExcept: OnlyExcept{Only: []string{"amazon-ebs.aws-ubuntu-16.04"}, Except: nil},
								},
							},
							{
								{
									PType:      "manifest",
									OnlyExcept: OnlyExcept{Only: nil, Except: []string{"amazon-ebs.aws-ubuntu-16.04"}},
								},
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:         "virtualbox-iso.ubuntu-1204",
					Prepared:     true,
					Builder:      emptyMockBuilder,
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType: "amazon-import",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: &MockPostProcessor{
										Config: MockConfig{
											NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
											NestedSlice:      []NestedMockConfig{},
										},
									},
								},
							},
						},
						{
							{
								PType: "manifest",
								PostProcessor: &HCL2PostProcessor{
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
				&packer.CoreBuild{
					Type:         "amazon-ebs.aws-ubuntu-16.04",
					Prepared:     true,
					Builder:      emptyMockBuilder,
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PType: "manifest",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: &MockPostProcessor{
										Config: MockConfig{
											NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
											NestedSlice:      []NestedMockConfig{},
										},
									},
								},
							},
						},
						{
							{
								PType: "amazon-import",
								PostProcessor: &HCL2PostProcessor{
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
			},
			false,
		},
		{"provisioner with only and except",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_onlyexcept.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204:  {Type: "virtualbox-iso", Name: "ubuntu-1204"},
					refAWSEBSUbuntu1604: {Type: "amazon-ebs", Name: "ubuntu-1604"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
							{
								SourceRef: SourceRef{Type: "amazon-ebs", Name: "ubuntu-1604"},
								LocalName: "aws-ubuntu-16.04",
							},
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType:      "shell",
								OnlyExcept: OnlyExcept{Only: []string{"virtualbox-iso.ubuntu-1204"}},
							},
							{
								PType:      "file",
								OnlyExcept: OnlyExcept{Except: []string{"virtualbox-iso.ubuntu-1204"}},
							},
							{
								PType:      "shell",
								OnlyExcept: OnlyExcept{Only: []string{"amazon-ebs.aws-ubuntu-16.04"}},
							},
							{
								PType:      "file",
								OnlyExcept: OnlyExcept{Except: []string{"amazon-ebs.aws-ubuntu-16.04"}},
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:     "virtualbox-iso.ubuntu-1204",
					Prepared: true,
					Builder:  emptyMockBuilder,
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PType: "shell",
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
										NestedSlice:      []NestedMockConfig{},
									},
								},
							},
						},
						{
							PType: "file",
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
										NestedSlice:      []NestedMockConfig{},
									},
								},
							},
						},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
				},
				&packer.CoreBuild{
					Type:     "amazon-ebs.aws-ubuntu-16.04",
					Prepared: true,
					Builder:  emptyMockBuilder,
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PType: "file",
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
										NestedSlice:      []NestedMockConfig{},
									},
								},
							},
						},
						{
							PType: "shell",
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{Tags: []MockTag{}},
										NestedSlice:      []NestedMockConfig{},
									},
								},
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
