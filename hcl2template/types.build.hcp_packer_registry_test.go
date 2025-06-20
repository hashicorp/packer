// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/builder/null"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

func Test_ParseHCPPackerRegistryBlock(t *testing.T) {
	t.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "hcp-par-test")

	defaultParser := getBasicParser()

	tests := []parseTest{
		{"build block level deprecated",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/build-block-ok-bucket.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name:              "bucket-slug",
						HCPPackerRegistry: &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "bucket-slug",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				wantDiag:         true,
				wantDiagHasError: false,
			},
		},
		{"bucket name OK multiple block",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/build-block-ok-multiple-build-block.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name:              "build1",
						HCPPackerRegistry: &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
					{
						Name: "build2",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "build1",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
				&packer.CoreBuild{
					BuildName:      "build2",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				wantDiag:         true,
				wantDiagHasError: false,
			},
		},
		{"bucket name OK multiple block second build block",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/build-block-ok-multiple-build-block-second-block.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name: "build1",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
					{
						Name:              "build2",
						HCPPackerRegistry: &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "build1",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
				&packer.CoreBuild{
					BuildName:      "build2",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				wantDiag:         true,
				wantDiagHasError: false,
			},
		},
		{"bucket name OK multiple block multiple declaration",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/build-block-error-multiple-hcp-declaration.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name:              "build1",
						HCPPackerRegistry: &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
					{
						Name:              "build2",
						HCPPackerRegistry: &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "build1",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
				&packer.CoreBuild{
					BuildName:      "build2",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				wantDiag:         true,
				wantDiagHasError: true,
			},
		},
		{"bucket_name left empty",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-empty-bucket.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				HCPPackerRegistry:       &HCPPackerRegistryBlock{Slug: ""},
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name: "bucket-slug",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "bucket-slug",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{Slug: ""},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"bucket_name as variable",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-variable-for-bucket-name.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				HCPPackerRegistry: &HCPPackerRegistryBlock{
					Slug: "variable-bucket-slug",
				},
				InputVariables: Variables{
					"bucket": &Variable{
						Name:   "bucket",
						Type:   cty.String,
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("variable-bucket-slug")}},
					},
				},
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
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					Type:           "virtualbox-iso.ubuntu-1204",
					Prepared:       true,
					Builder:        emptyMockBuilder,
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					BuilderType:    "virtualbox-iso",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{

				wantBlock:        &HCPPackerRegistryBlock{Slug: "variable-bucket-slug"},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"bucket_labels and build_labels as variables",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-variables-for-labels.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				HCPPackerRegistry: &HCPPackerRegistryBlock{
					Slug:         "bucket-slug",
					BucketLabels: map[string]string{"team": "development"},
					BuildLabels:  map[string]string{"packageA": "v3.17.5", "packageZ": "v0.6"},
				},
				InputVariables: Variables{
					"bucket_labels": &Variable{
						Name:   "bucket_labels",
						Type:   cty.Map(cty.String),
						Values: []VariableAssignment{{From: "default", Value: cty.MapVal(map[string]cty.Value{"team": cty.StringVal("development")})}},
					},
					"build_labels": &Variable{
						Name: "build_labels",
						Type: cty.Map(cty.String),
						Values: []VariableAssignment{{
							From: "default",
							Value: cty.MapVal(map[string]cty.Value{
								"packageA": cty.StringVal("v3.17.5"),
								"packageZ": cty.StringVal("v0.6"),
							})}},
					},
				},
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
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					Type:           "virtualbox-iso.ubuntu-1204",
					Prepared:       true,
					Builder:        emptyMockBuilder,
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					BuilderType:    "virtualbox-iso",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock: &HCPPackerRegistryBlock{
					Slug:         "bucket-slug",
					BucketLabels: map[string]string{"team": "development"},
					BuildLabels:  map[string]string{"packageA": "v3.17.5", "packageZ": "v0.6"},
				},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"invalid hcp_packer_registry config",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-invalid.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
			},
			true, true,
			nil,
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"long hcp_packer_registry.description",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-long-description.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name: "bucket-slug",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"bucket name too short",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-short-bucket.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name: "bucket-slug",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"bucket name too long",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-long-bucket.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name: "bucket-slug",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"bucket name invalid chars",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-invalid-bucket.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name: "bucket-slug",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"bucket name OK",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-ok-bucket.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				HCPPackerRegistry:       &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name: "bucket-slug",
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "bucket-slug",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				wantDiag:         false,
				wantDiagHasError: false,
			},
		},
		{"top level and build block",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/top-level-and-build-block.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
				HCPPackerRegistry:       &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
				Builds: Builds{
					{
						Name:              "bucket-slug",
						HCPPackerRegistry: &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
						Sources: []SourceUseBlock{
							{
								SourceRef: refNull,
							},
						},
					},
				},
			},
			false, false,
			[]*packer.CoreBuild{
				&packer.CoreBuild{
					BuildName:      "bucket-slug",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
					SensitiveVars:  []string{},
				},
			},
			false,
			&getHCPPackerRegistry{
				wantBlock:        &HCPPackerRegistryBlock{Slug: "ok-Bucket-name-1"},
				wantDiag:         true,
				wantDiagHasError: true,
			},
		},
	}
	testParse(t, tests)
}
