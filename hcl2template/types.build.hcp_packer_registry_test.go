// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/null"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

func Test_ParseHCPPackerRegistryBlock(t *testing.T) {
	t.Setenv("HCP_PACKER_BUILD_FINGERPRINT", "hcp-par-test")

	defaultParser := getBasicParser()

	tests := []parseTest{
		{"bucket_name as variable",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/variable-for-bucket_name.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
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
						HCPPackerRegistry: &HCPPackerRegistryBlock{
							Slug: "variable-bucket-slug",
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
					Type:           "virtualbox-iso.ubuntu-1204",
					Prepared:       true,
					Builder:        emptyMockBuilder,
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					BuilderType:    "virtualbox-iso",
				},
			},
			false,
		},
		{"bucket_labels and build_labels as variables",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/variables-for-labels.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "hcp_par"),
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
						HCPPackerRegistry: &HCPPackerRegistryBlock{
							Slug:         "bucket-slug",
							BucketLabels: map[string]string{"team": "development"},
							BuildLabels:  map[string]string{"packageA": "v3.17.5", "packageZ": "v0.6"},
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
					Type:           "virtualbox-iso.ubuntu-1204",
					Prepared:       true,
					Builder:        emptyMockBuilder,
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					BuilderType:    "virtualbox-iso",
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
				Sources: map[SourceRef]SourceBlock{
					refNull: {
						Type: "null",
						Name: "test",
						block: &hcl.Block{
							Type: "source",
						},
					},
				},
			},
			true, true,
			nil,
			false,
		},
		{"bucket name too short",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/short_bucket.pkr.hcl", nil, nil},
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
			},
			true, true,
			nil,
			false,
		},
		{"bucket name too long",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/long_bucket.pkr.hcl", nil, nil},
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
			},
			true, true,
			nil,
			false,
		},
		{"bucket name invalid chars",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/invalid_bucket.pkr.hcl", nil, nil},
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
			},
			true, true,
			nil,
			false,
		},
		{"bucket name OK",
			defaultParser,
			parseTestArgs{"testdata/hcp_par/ok_bucket.pkr.hcl", nil, nil},
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
			[]packersdk.Build{
				&packer.CoreBuild{
					BuildName:      "bucket-slug",
					Type:           "null.test",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
					BuilderType:    "null",
				},
			},
			false,
		},
	}
	testParse(t, tests)
}
