package hcl2template

import (
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/internal/packer_registry"
	"github.com/hashicorp/packer/packer"
)

func Test_ParseHCPPackerRegistryBlock(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"complete working build with hcp_packer_registry block",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/complete.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204:  {Type: "virtualbox-iso", Name: "ubuntu-1204"},
					refAWSEBSUbuntu1604: {Type: "amazon-ebs", Name: "ubuntu-1604"},
				},
				Builds: Builds{
					&BuildBlock{
						Name: "bucket-slug",
						HCPPackerRegistry: &HCPPackerRegistryBlock{
							Description: "Some description\n",
							Labels:      map[string]string{"foo": "bar"},
						},
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
							{
								SourceRef: SourceRef{Type: "amazon-ebs", Name: "ubuntu-1604"},
								LocalName: "aws-ubuntu-16.04",
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					BuildName: "bucket-slug",
					Type:      "virtualbox-iso.ubuntu-1204",
					Prepared:  true,
					Builder: &packer.RegistryBuilder{
						Name:    "virtualbox-iso.ubuntu-1204",
						Builder: emptyMockBuilder,
						ArtifactMetadataPublisher: &packer_registry.Bucket{
							Slug:        "bucket-slug",
							Description: "Some description\n",
							Labels:      map[string]string{"foo": "bar"},
							Iteration: &packer_registry.Iteration{
								Fingerprint: "ignored-fingerprint", // this will be different everytime so it's ignored
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "virtualbox-iso.ubuntu-1204",
									ArtifactMetadataPublisher: &packer_registry.Bucket{
										Slug:        "bucket-slug",
										Description: "Some description\n",
										Labels:      map[string]string{"foo": "bar"},
										Iteration: &packer_registry.Iteration{
											Fingerprint: "ignored-fingerprint",
										},
									},
								},
							},
						},
					},
				},
				&packer.CoreBuild{
					BuildName: "bucket-slug",
					Type:      "amazon-ebs.aws-ubuntu-16.04",
					Prepared:  true,
					Builder: &packer.RegistryBuilder{
						Name:    "amazon-ebs.aws-ubuntu-16.04",
						Builder: emptyMockBuilder,
						ArtifactMetadataPublisher: &packer_registry.Bucket{
							Slug:        "bucket-slug",
							Description: "Some description\n",
							Labels:      map[string]string{"foo": "bar"},
							Iteration: &packer_registry.Iteration{
								Fingerprint: "ignored-fingerprint", // this will be different everytime so it's ignored
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "amazon-ebs.aws-ubuntu-16.04",
									ArtifactMetadataPublisher: &packer_registry.Bucket{
										Slug:        "bucket-slug",
										Description: "Some description\n",
										Labels:      map[string]string{"foo": "bar"},
										Iteration: &packer_registry.Iteration{
											Fingerprint: "ignored-fingerprint",
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
		{"set slug in hcp_packer_registry block",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/slug.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Name: "bucket-slug",
						HCPPackerRegistry: &HCPPackerRegistryBlock{
							Slug:        "real-bucket-slug",
							Description: "Some description\n",
							Labels:      map[string]string{"foo": "bar"},
						},
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					BuildName: "bucket-slug",
					Type:      "virtualbox-iso.ubuntu-1204",
					Prepared:  true,
					Builder: &packer.RegistryBuilder{
						Name:    "virtualbox-iso.ubuntu-1204",
						Builder: emptyMockBuilder,
						ArtifactMetadataPublisher: &packer_registry.Bucket{
							Slug:        "real-bucket-slug",
							Description: "Some description\n",
							Labels:      map[string]string{"foo": "bar"},
							Iteration: &packer_registry.Iteration{
								Fingerprint: "ignored-fingerprint", // this will be different everytime so it's ignored
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "virtualbox-iso.ubuntu-1204",
									ArtifactMetadataPublisher: &packer_registry.Bucket{
										Slug:        "real-bucket-slug",
										Description: "Some description\n",
										Labels:      map[string]string{"foo": "bar"},
										Iteration: &packer_registry.Iteration{
											Fingerprint: "ignored-fingerprint",
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
		{"use build description",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/build-description.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Description: "Some build description\n",
						HCPPackerRegistry: &HCPPackerRegistryBlock{
							Slug: "bucket-slug",
						},
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
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
					Builder: &packer.RegistryBuilder{
						Name:    "virtualbox-iso.ubuntu-1204",
						Builder: emptyMockBuilder,
						ArtifactMetadataPublisher: &packer_registry.Bucket{
							Slug:        "bucket-slug",
							Description: "Some build description\n",
							Iteration: &packer_registry.Iteration{
								Fingerprint: "ignored-fingerprint", // this will be different everytime so it's ignored
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "virtualbox-iso.ubuntu-1204",
									ArtifactMetadataPublisher: &packer_registry.Bucket{
										Slug:        "bucket-slug",
										Description: "Some build description\n",
										Iteration: &packer_registry.Iteration{
											Fingerprint: "ignored-fingerprint",
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
		{"override build description with hcp_packer_registry description",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/override-build-description.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Description: "Some build description\n",
						HCPPackerRegistry: &HCPPackerRegistryBlock{
							Slug:        "bucket-slug",
							Description: "Some override description\n",
						},
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
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
					Builder: &packer.RegistryBuilder{
						Name:    "virtualbox-iso.ubuntu-1204",
						Builder: emptyMockBuilder,
						ArtifactMetadataPublisher: &packer_registry.Bucket{
							Slug:        "bucket-slug",
							Description: "Some override description\n",
							Iteration: &packer_registry.Iteration{
								Fingerprint: "ignored-fingerprint", // this will be different everytime so it's ignored
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "virtualbox-iso.ubuntu-1204",
									ArtifactMetadataPublisher: &packer_registry.Bucket{
										Slug:        "bucket-slug",
										Description: "Some override description\n",
										Iteration: &packer_registry.Iteration{
											Fingerprint: "ignored-fingerprint",
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
		{"duplicate hcp_packer_registry blocks",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/duplicate.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
			},
			true, true,
			nil,
			false,
		},
		{"empty hcp_packer_registry block",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/empty.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refVBIsoUbuntu1204: {Type: "virtualbox-iso", Name: "ubuntu-1204"},
				},
				Builds: Builds{
					&BuildBlock{
						Name:              "bucket-slug",
						HCPPackerRegistry: &HCPPackerRegistryBlock{},
						Sources: []SourceUseBlock{
							{
								SourceRef: refVBIsoUbuntu1204,
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					BuildName: "bucket-slug",
					Type:      "virtualbox-iso.ubuntu-1204",
					Prepared:  true,
					Builder: &packer.RegistryBuilder{
						Name:    "virtualbox-iso.ubuntu-1204",
						Builder: emptyMockBuilder,
						ArtifactMetadataPublisher: &packer_registry.Bucket{
							Slug: "bucket-slug",
							Iteration: &packer_registry.Iteration{
								Fingerprint: "ignored-fingerprint", // this will be different everytime so it's ignored
							},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "virtualbox-iso.ubuntu-1204",
									ArtifactMetadataPublisher: &packer_registry.Bucket{
										Slug: "bucket-slug",
										Iteration: &packer_registry.Iteration{
											Fingerprint: "ignored-fingerprint",
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
		{"invalid hcp_packer_registry config",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/invalid.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
			},
			true, true,
			nil,
			false,
		},
		{"long hcp_packer_registry.description",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/long-description.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
			},
			true, true,
			nil,
			false,
		},
	}
	testParse(t, tests)
}
