// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

func TestPackerConfig_required_plugin_parse(t *testing.T) {

	tests := []struct {
		name           string
		cfg            PackerConfig
		requirePlugins string
		restOfTemplate string
		wantDiags      bool
		wantConfig     PackerConfig
	}{
		{"required_plugin", PackerConfig{parser: getBasicParser()}, `
		packer {
			required_plugins {
				amazon = {
					source  = "github.com/hashicorp/amazon"
					version = "~> v1.2.3"
				}
			}
		} `, `
		source "amazon-ebs" "example" {
		}
		`, false, PackerConfig{
			Packer: struct {
				VersionConstraints []VersionConstraint
				RequiredPlugins    []*RequiredPlugins
			}{
				RequiredPlugins: []*RequiredPlugins{
					{RequiredPlugins: map[string]*RequiredPlugin{
						"amazon": {
							Name:   "amazon",
							Source: "github.com/hashicorp/amazon",
							Type:   &addrs.Plugin{Source: "github.com/hashicorp/amazon"},
							Requirement: VersionConstraint{
								Required: mustVersionConstraints(version.NewConstraint("~> v1.2.3")),
							},
						},
					}},
				},
			},
		}},
		{"required_plugin_forked_no_redirect", PackerConfig{parser: getBasicParser()}, `
		packer {
			required_plugins {
				amazon = {
					source  = "github.com/azr/amazon"
					version = "~> v1.2.3"
				}
			}
		} `, `
		source "amazon-chroot" "example" {
		}
		`, false, PackerConfig{
			Packer: struct {
				VersionConstraints []VersionConstraint
				RequiredPlugins    []*RequiredPlugins
			}{
				RequiredPlugins: []*RequiredPlugins{
					{RequiredPlugins: map[string]*RequiredPlugin{
						"amazon": {
							Name:   "amazon",
							Source: "github.com/azr/amazon",
							Type:   &addrs.Plugin{Source: "github.com/azr/amazon"},
							Requirement: VersionConstraint{
								Required: mustVersionConstraints(version.NewConstraint("~> v1.2.3")),
							},
						},
					}},
				},
			},
		}},
		{"required_plugin_forked", PackerConfig{
			parser: getBasicParser(func(p *Parser) {})}, `
		packer {
			required_plugins {
				amazon = {
					source  = "github.com/azr/amazon"
					version = "~> v1.2.3"
				}
			}
		} `, `
		source "amazon-chroot" "example" {
		}
		`, false, PackerConfig{
			Packer: struct {
				VersionConstraints []VersionConstraint
				RequiredPlugins    []*RequiredPlugins
			}{
				RequiredPlugins: []*RequiredPlugins{
					{RequiredPlugins: map[string]*RequiredPlugin{
						"amazon": {
							Name:   "amazon",
							Source: "github.com/azr/amazon",
							Type:   &addrs.Plugin{Source: "github.com/azr/amazon"},
							Requirement: VersionConstraint{
								Required: mustVersionConstraints(version.NewConstraint("~> v1.2.3")),
							},
						},
					}},
				},
			},
		}},
		{"missing-required-plugin-for-pre-defined-builder", PackerConfig{
			parser: getBasicParser(func(p *Parser) {})},
			`
			packer {
			}`, `
			# amazon-ebs is mocked in getBasicParser()
			source "amazon-ebs" "example" {
			}
			`,
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: nil,
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			file, diags := cfg.parser.ParseHCL([]byte(tt.requirePlugins), "required_plugins.pkr.hcl")
			if len(diags) > 0 {
				t.Fatal(diags)
			}
			if diags := cfg.decodeRequiredPluginsBlock(file); len(diags) > 0 {
				t.Fatal(diags)
			}

			_, diags = cfg.parser.ParseHCL([]byte(tt.restOfTemplate), "rest.pkr.hcl")
			if len(diags) > 0 {
				t.Fatal(diags)
			}
			if diff := cmp.Diff(tt.wantConfig, cfg, cmpOpts...); diff != "" {
				t.Errorf("PackerConfig.inferImplicitRequiredPluginFromBlocks() unexpected PackerConfig: %v", diff)
			}
		})
	}
}
