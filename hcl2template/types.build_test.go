// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"path/filepath"
	"testing"

	. "github.com/hashicorp/packer/hcl2template/internal"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
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
			[]*packer.CoreBuild{},
			true,
			nil,
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
			nil,
		},
		{"nonexistent provisioner",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_nonexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
				Builds: Builds{
					&BuildBlock{
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "nonexistent",
							},
						},
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
			},
			true, true,
			[]*packer.CoreBuild{&packer.CoreBuild{
				Provisioners:  []packer.CoreBuildProvisioner{},
				SensitiveVars: []string{},
			}},
			false,
			nil,
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
			[]*packer.CoreBuild{&packer.CoreBuild{
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
			nil,
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
			[]*packer.CoreBuild{&packer.CoreBuild{
				SensitiveVars: []string{},
			}},
			false,
			nil,
		},
		{"nonexistent post-processor",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_nonexistent.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
				Builds: Builds{
					&BuildBlock{
						PostProcessorsLists: [][]*PostProcessorBlock{
							{
								{
									PType: "nonexistent",
								},
							},
						},
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
			},
			true, true,
			[]*packer.CoreBuild{&packer.CoreBuild{
				PostProcessors: [][]packer.CoreBuildPostProcessor{},
				SensitiveVars:  []string{},
			}},
			true,
			nil,
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
			[]*packer.CoreBuild{},
			false,
			nil,
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
			[]*packer.CoreBuild{},
			true,
			nil,
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
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					Type:          "virtualbox-iso.ubuntu-1204",
					BuilderType:   "virtualbox-iso",
					Prepared:      true,
					Builder:       emptyMockBuilder,
					Provisioners:  []packer.CoreBuildProvisioner{},
					SensitiveVars: []string{},
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
					Type:          "amazon-ebs.aws-ubuntu-16.04",
					BuilderType:   "amazon-ebs",
					Prepared:      true,
					Builder:       emptyMockBuilder,
					Provisioners:  []packer.CoreBuildProvisioner{},
					SensitiveVars: []string{},
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
			nil,
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
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					Type:          "virtualbox-iso.ubuntu-1204",
					BuilderType:   "virtualbox-iso",
					Prepared:      true,
					Builder:       emptyMockBuilder,
					SensitiveVars: []string{},
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
					Type:          "amazon-ebs.aws-ubuntu-16.04",
					BuilderType:   "amazon-ebs",
					Prepared:      true,
					Builder:       emptyMockBuilder,
					SensitiveVars: []string{},
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
			nil,
		},
		{"provisioner with packer_version interpolation",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_packer_version_interpolation.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType: "shell",
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					Type:          "virtualbox-iso.ubuntu-1204",
					BuilderType:   "virtualbox-iso",
					Prepared:      true,
					Builder:       emptyMockBuilder,
					SensitiveVars: []string{},
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PType: "shell",
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{
											Tags:        []MockTag{},
											SliceString: []string{lockedVersion},
										},
										NestedSlice: []NestedMockConfig{},
									},
								},
							},
						},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
				},
			},
			false,
			nil,
		},
		{"variable interpolation for build name and description",
			defaultParser,
			parseTestArgs{"testdata/build/variables.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				InputVariables: Variables{
					"name": &Variable{
						Name:   "name",
						Type:   cty.String,
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("build-name")}},
					},
				},
				LocalVariables: Variables{
					"description": &Variable{
						Name:   "description",
						Type:   cty.String,
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("This is the description for build-name.")}},
					},
				},
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Name:        "build-name",
						Description: "This is the description for build-name.",
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "build-name",
					Type:           "virtualbox-iso.ubuntu-1204",
					BuilderType:    "virtualbox-iso",
					Prepared:       true,
					Builder:        emptyMockBuilder,
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					SensitiveVars:  []string{},
				},
			},
			false,
			nil,
		},
		{"invalid variable for build name",
			defaultParser,
			parseTestArgs{"testdata/build/invalid_build_name_variable.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				InputVariables:          Variables{},
				Builds:                  nil,
			},
			true, true,
			[]*packer.CoreBuild{},
			false,
			nil,
		},
		{"use build.name in post-processor block",
			defaultParser,
			parseTestArgs{"testdata/build/post-processor_build_name_interpolation.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Name: "test-build",
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
						PostProcessorsLists: [][]*PostProcessorBlock{
							{
								{
									PName: "test-build",
									PType: "manifest",
								},
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:     "test-build",
					Type:          "virtualbox-iso.ubuntu-1204",
					BuilderType:   "virtualbox-iso",
					Prepared:      true,
					Builder:       emptyMockBuilder,
					Provisioners:  []packer.CoreBuildProvisioner{},
					SensitiveVars: []string{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PName: "test-build",
								PType: "manifest",
								PostProcessor: &HCL2PostProcessor{
									PostProcessor: &MockPostProcessor{
										Config: MockConfig{
											NestedMockConfig: NestedMockConfig{
												Tags:        []MockTag{},
												SliceString: []string{lockedVersion, "test-build"},
											},
											NestedSlice: []NestedMockConfig{},
										},
									},
								},
							},
						},
					},
				},
			},
			false,
			nil,
		},
		{"use build.name in provisioner block",
			defaultParser,
			parseTestArgs{"testdata/build/provisioner_build_name_interpolation.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "build"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Name: "build-name-test",
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PName: "build-name-test",
								PType: "shell",
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:     "build-name-test",
					Type:          "virtualbox-iso.ubuntu-1204",
					BuilderType:   "virtualbox-iso",
					Prepared:      true,
					Builder:       emptyMockBuilder,
					SensitiveVars: []string{},
					Provisioners: []packer.CoreBuildProvisioner{
						{
							PName: "build-name-test",
							PType: "shell",
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{
											Tags:        []MockTag{},
											SliceString: []string{"build-name-test"},
										},
										NestedSlice: []NestedMockConfig{},
									},
								},
							},
						},
					},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
				},
			},
			false,
			nil,
		},
	}
	testParse(t, tests)
}
