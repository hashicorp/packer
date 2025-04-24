package registry

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/packer/hcl2template"
)

func TestNewRegisterProperBuildName(t *testing.T) {
	cases := map[string]struct {
		expectedBuilds       []string
		expectErr            bool
		diagsSummaryContains string
		builds               hcl2template.Builds
	}{
		"single build block with single source": {
			expectErr:      false,
			expectedBuilds: []string{"docker.ubuntu"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
			},
		},
		"single build block with name and with single source": {
			expectErr:      false,
			expectedBuilds: []string{"docker.ubuntu"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Name: "my-build-block",
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
			},
		},
		"single build block with 2 sources": {
			expectErr:      false,
			expectedBuilds: []string{"docker.alpine", "docker.ubuntu"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "alpine",
							},
						},
					},
				},
			},
		},
		"single build block with 3 sources": {
			expectErr:      false,
			expectedBuilds: []string{"docker.alpine", "docker.ubuntu", "docker.arch"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "alpine",
							},
						},
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "arch",
							},
						},
					},
				},
			},
		},
		"single build block with name and multiple sources": {
			expectErr:      false,
			expectedBuilds: []string{"docker.alpine", "docker.ubuntu"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Name: "my-build-block",
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "alpine",
							},
						},
					},
				},
			},
		},
		"single build block with multiple identical sources create conflict": {
			expectErr:            true,
			diagsSummaryContains: "conflict",
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
			},
		},
		"multiple build block with different source": {
			expectErr:      false,
			expectedBuilds: []string{"docker.alpine", "docker.ubuntu"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "alpine",
							},
						},
					},
				},
			},
		},
		"multiple build block with same source create conflict": {
			expectErr:            true,
			diagsSummaryContains: "conflict",
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "alpine",
							},
						},
					},
				},
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "alpine",
							},
						},
					},
				},
			},
		},
		"multiple build block with same source but with different build name": {
			expectErr:      false,
			expectedBuilds: []string{"build1.docker.ubuntu", "build2.docker.ubuntu"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Name: "build1",
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
				&hcl2template.BuildBlock{
					Name: "build2",
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
			},
		},
		"multiple build block with same source but with only one declared build name": {
			expectErr:      false,
			expectedBuilds: []string{"docker.ubuntu", "build.docker.ubuntu"},
			builds: hcl2template.Builds{
				&hcl2template.BuildBlock{
					Name: "build",
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
				&hcl2template.BuildBlock{
					Sources: []hcl2template.SourceUseBlock{
						{
							SourceRef: hcl2template.SourceRef{
								Type: "docker",
								Name: "ubuntu",
							},
						},
					},
				},
			},
		},
	}

	for desc, tc := range cases {
		t.Run(desc, func(t *testing.T) {

			config := &hcl2template.PackerConfig{
				Builds: tc.builds,
			}

			registry := HCLRegistry{
				configuration: config,
				bucket: &Bucket{
					Name:    "test-bucket-" + desc,
					Version: &Version{},
				},
				buildNames: map[string]struct{}{},
			}

			diags := registry.registerAllComponents()
			if tc.diagsSummaryContains != "" {

				containsMsg := false
				for _, diag := range diags {
					if strings.Contains(diag.Summary, tc.diagsSummaryContains) {
						containsMsg = true
					}
				}
				if !containsMsg {
					t.Fatalf("diagnostics should contains '%s' in summary", tc.diagsSummaryContains)
				}
			}
			if !tc.expectErr {
				if diags.HasErrors() {
					t.Fatalf("should not report error diagnostic: %v", diags)
				}
			}
			if tc.expectErr {
				if !diags.HasErrors() {
					t.Fatal("should report error in this case")
				}
				return
			}

			actualExpectedBuilds := registry.bucket.Version.expectedBuilds

			slices.Sort(tc.expectedBuilds)
			slices.Sort(actualExpectedBuilds)

			if !reflect.DeepEqual(tc.expectedBuilds, actualExpectedBuilds) {
				t.Fatalf("expectedBuilds registered: %v, got: %v", tc.expectedBuilds, actualExpectedBuilds)
			}
		})
	}
}
