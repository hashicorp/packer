package command

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/internal/registry"
	"github.com/hashicorp/packer/packer"
)

type registryTestArgs struct {
	name           string
	inputFilePath  string
	expectedBuilds []packersdk.Build
	expectError    bool
	envvars        map[string]string
}

var registryCmpOpts = cmp.Options{
	cmpopts.IgnoreFields(
		registry.Iteration{},
		"Fingerprint",
	),
	cmpopts.IgnoreFields(
		packer.CoreBuild{},
		"TemplatePath",
		"Variables",
	),
	cmpopts.IgnoreUnexported(
		hcl2template.PackerConfig{},
		hcl2template.Variable{},
		hcl2template.SourceBlock{},
		hcl2template.DatasourceBlock{},
		hcl2template.ProvisionerBlock{},
		hcl2template.PostProcessorBlock{},
		packer.CoreBuild{},
		hcl2template.HCL2Provisioner{},
		hcl2template.HCL2PostProcessor{},
		packer.CoreBuildPostProcessor{},
		packer.CoreBuildProvisioner{},
		packer.CoreBuildPostProcessor{},
		file.Builder{},
		registry.Bucket{},
		registry.Iteration{},
		packer.RegistryBuilder{},
	),
}

func TestRegistrySetup(t *testing.T) {
	tests := []registryTestArgs{
		{
			"HCL2 - hcp_registry_block and multiple sources",
			"test-fixtures/hcp/multiple_sources.pkr.hcl",
			[]packersdk.Build{
				&packer.CoreBuild{
					BuildName: "bucket-slug",
					Type:      "file.test",
					Prepared:  true,
					Builder: &packer.RegistryBuilder{
						Name:    "file.test",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "bucket-slug",
							Description:                    "Some description\n",
							BucketLabels:                   map[string]string{"foo": "bar"},
							BuildLabels:                    map[string]string{"python_version": "3.0"},
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file.test",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "bucket-slug",
										Description:                    "Some description\n",
										BucketLabels:                   map[string]string{"foo": "bar"},
										BuildLabels:                    map[string]string{"python_version": "3.0"},
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
									},
								},
							},
						},
					},
				},
				&packer.CoreBuild{
					BuildName: "bucket-slug",
					Type:      "file.other",
					Prepared:  true,
					Builder: &packer.RegistryBuilder{
						Name:    "file.other",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "bucket-slug",
							Description:                    "Some description\n",
							BucketLabels:                   map[string]string{"foo": "bar"},
							BuildLabels:                    map[string]string{"python_version": "3.0"},
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file.other",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "bucket-slug",
										Description:                    "Some description\n",
										BucketLabels:                   map[string]string{"foo": "bar"},
										BuildLabels:                    map[string]string{"python_version": "3.0"},
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
									},
								},
							},
						},
					},
				},
			},
			false,
			map[string]string{},
		},
		{
			"HCL2 - set slug in hcp packer registry block",
			"test-fixtures/hcp/slug.pkr.hcl",
			[]packersdk.Build{
				&packer.CoreBuild{
					BuildName: "bucket-slug",
					Type:      "file.test",
					Prepared:  true,
					Builder: &packer.RegistryBuilder{
						Name:    "file.test",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "real-bucket-slug",
							Description:                    "Some description\n",
							BucketLabels:                   map[string]string{"foo": "bar"},
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file.test",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "real-bucket-slug",
										Description:                    "Some description\n",
										BucketLabels:                   map[string]string{"foo": "bar"},
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
									},
								},
							},
						},
					},
				},
			},
			false,
			map[string]string{},
		},
		{
			"HCL2 - hcp - use build description",
			"test-fixtures/hcp/build-description.pkr.hcl",
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:      "file.test",
					Prepared:  true,
					BuildName: "bucket-slug",
					Builder: &packer.RegistryBuilder{
						Name:    "file.test",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "bucket-slug",
							Description:                    "Some build description",
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file.test",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "bucket-slug",
										Description:                    "Some build description",
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
									},
								},
							},
						},
					},
				},
			},
			false,
			map[string]string{},
		},
		{
			"HCL2 - override build description with hcp packer registry description",
			"test-fixtures/hcp/override-build-description.pkr.hcl",
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:     "file.test",
					Prepared: true,
					Builder: &packer.RegistryBuilder{
						Name:    "file.test",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "bucket-slug",
							Description:                    "Some override description",
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file.test",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "bucket-slug",
										Description:                    "Some override description",
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
									},
								},
							},
						},
					},
				},
			},
			false,
			map[string]string{},
		},
		{
			"HCL2 - deprecated labels in hcp packer registry block",
			"test-fixtures/hcp/deprecated_labels.pkr.hcl",
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:     "file.test",
					Prepared: true,
					Builder: &packer.RegistryBuilder{
						Name:    "file.test",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "bucket-slug",
							BucketLabels:                   map[string]string{"foo": "bar"},
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file.test",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "bucket-slug",
										BucketLabels:                   map[string]string{"foo": "bar"},
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
									},
								},
							},
						},
					},
				},
			},
			true,
			map[string]string{},
		},
		{
			"HCL2 - duplicate hcp_registry blocks",
			"test-fixtures/hcp/duplicate.pkr.hcl",
			nil,
			true,
			map[string]string{},
		},
		{
			"HCL2 - two build blocks with hcp_registry",
			"test-fixtures/hcp/dup_build_blocks.pkr.hcl",
			nil,
			true,
			map[string]string{},
		},
		{
			"HCL2/JSON - hcp enabled build",
			"test-fixtures/hcp/hcp_normal.pkr.json",
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:      "file.test",
					Prepared:  true,
					BuildName: "test-file",
					Builder: &packer.RegistryBuilder{
						Name:    "file.test",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "bucket-slug",
							Description:                    "Some build description",
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
							BucketLabels:                   map[string]string{},
							BuildLabels:                    map[string]string{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file.test",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "bucket-slug",
										Description:                    "Some build description",
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
										BucketLabels:                   map[string]string{},
										BuildLabels:                    map[string]string{},
									},
								},
							},
						},
					},
				},
			},
			false,
			map[string]string{
				"HCP_PACKER_REGISTRY":    "1",
				"HCP_PACKER_BUCKET_NAME": "bucket-slug",
			},
		},
		{
			"Legacy JSON - hcp enabled build",
			"test-fixtures/hcp/hcp_build.json",
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:        "file",
					Prepared:    false,
					BuildName:   "",
					BuilderType: "file",
					BuilderConfig: map[string]interface{}{
						"content": " ",
						"target":  "output",
					},
					Builder: &packer.RegistryBuilder{
						Name:    "file",
						Builder: &file.Builder{},
						ArtifactMetadataPublisher: &registry.Bucket{
							Slug:                           "bucket-slug",
							Description:                    "",
							Iteration:                      &registry.Iteration{},
							SourceImagesToParentIterations: map[string]registry.ParentIteration{},
							BucketLabels:                   map[string]string{},
							BuildLabels:                    map[string]string{},
						},
					},
					Provisioners: []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{
						{
							{
								PostProcessor: &packer.RegistryPostProcessor{
									BuilderType: "file",
									ArtifactMetadataPublisher: &registry.Bucket{
										Slug:                           "bucket-slug",
										Description:                    "",
										Iteration:                      &registry.Iteration{},
										SourceImagesToParentIterations: map[string]registry.ParentIteration{},
										BucketLabels:                   map[string]string{},
										BuildLabels:                    map[string]string{},
									},
								},
							},
						},
					},
				},
			},
			false,
			map[string]string{
				"HCP_PACKER_REGISTRY":    "1",
				"HCP_PACKER_BUCKET_NAME": "bucket-slug",
			},
		},
	}

	t.Setenv("HCP_CLIENT_ID", "test")
	t.Setenv("HCP_CLIENT_SECRET", "test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runRegistryTest(t, tt)
		})
	}
}

func runRegistryTest(t *testing.T, args registryTestArgs) {
	for evar, val := range args.envvars {
		t.Setenv(evar, val)
	}

	defaultMeta := TestMetaFile(t)

	c := &BuildCommand{
		Meta: defaultMeta,
	}

	cla := &BuildArgs{
		MetaArgs: MetaArgs{
			Path: args.inputFilePath,
		},
	}

	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		t.Fatalf("failed to get packer config")
	}

	diags := packerStarter.Initialize(packer.InitializeOptions{})
	if diagCheck(diags, t) {
		return
	}

	diags = TrySetupHCP(packerStarter)
	if diagCheck(diags, t) {
		if !args.expectError {
			t.Errorf("SetupRegistry unexpectedly failed")
		}
		return
	}

	builds, diags := packerStarter.GetBuilds(packer.GetBuildsOptions{
		Only:    cla.Only,
		Except:  cla.Except,
		Debug:   cla.Debug,
		Force:   cla.Force,
		OnError: cla.OnError,
	})
	if diagCheck(diags, t) {
		if !args.expectError {
			t.Errorf("SetupRegistry unexpectedly failed")
		}
		return
	}

	diff := cmp.Diff(builds, args.expectedBuilds, registryCmpOpts...)
	if diff != "" {
		t.Error(diff)
	}
}

func severityString(sev hcl.DiagnosticSeverity) string {
	switch sev {
	case hcl.DiagInvalid:
		return "UNKNOWN"
	case hcl.DiagError:
		return "ERROR"
	case hcl.DiagWarning:
		return "WARNING"
	}
	panic("unknown severity")
}

// diagCheck errors for each diagnostic received and returns if a diag was processed
func diagCheck(diags hcl.Diagnostics, t *testing.T) bool {
	if len(diags) == 0 {
		return false
	}

	for _, d := range diags {
		t.Logf(
			"%s: %s - %s",
			severityString(d.Severity),
			d.Summary,
			d.Detail,
		)
	}
	return true
}
