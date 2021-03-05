package hcl2template

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

func TestPackerConfig_required_plugin_parse(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []struct {
		name           string
		cfg            PackerConfig
		requirePlugins string
		restOfTemplate string
		wantDiags      bool
		wantConfig     PackerConfig
	}{
		{"required_plugin", PackerConfig{parser: defaultParser}, `
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
							Type:   &addrs.Plugin{"github.com", "hashicorp", "amazon"},
							Requirement: VersionConstraint{
								Required: mustVersionConstraints(version.NewConstraint("~> v1.2.3")),
							},
							PluginDependencyReason: PluginDependencyExplicit,
						},
					}},
				},
			},
		}},
		{"missing-required-plugin-for-builder", PackerConfig{
			parser: getBasicParser(func(p *Parser) {
				p.PluginConfig.BuilderRedirects = map[string]string{
					"amazon-ebs": "github.com/hashicorp/amazon",
				}
			},
			)},
			`
			packer {
			}`, `
			source "amazon-ebs" "example" {
			}
			`,
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{RequiredPlugins: map[string]*RequiredPlugin{
							"amazon": {
								Name:   "amazon",
								Source: "github.com/hashicorp/amazon",
								Type:   &addrs.Plugin{"github.com", "hashicorp", "amazon"},
								Requirement: VersionConstraint{
									Required: nil,
								},
								PluginDependencyReason: PluginDependencyImplicit,
							},
						}},
					},
				},
			}},
		{"missing-required-plugin-for-provisioner", PackerConfig{
			parser: getBasicParser(func(p *Parser) {
				p.PluginConfig.ProvisionerRedirects = map[string]string{
					"ansible-local": "github.com/ansible/ansible",
				}
			},
			)},
			`
			packer {
			}`, `
			build {
				provisioner "ansible-local" {}
			}
			`,
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{RequiredPlugins: map[string]*RequiredPlugin{
							"ansible": {
								Name:   "ansible",
								Source: "github.com/ansible/ansible",
								Type:   &addrs.Plugin{"github.com", "ansible", "ansible"},
								Requirement: VersionConstraint{
									Required: nil,
								},
								PluginDependencyReason: PluginDependencyImplicit,
							},
						}},
					},
				},
			}},
		{"missing-required-plugin-for-post-processor", PackerConfig{
			parser: getBasicParser(func(p *Parser) {
				p.PluginConfig.PostProcessorRedirects = map[string]string{
					"docker-push": "github.com/hashicorp/docker",
				}
			},
			)},
			`
			packer {
			}`, `
			build {
				post-processor "docker-push" {}
			}
			`,
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{RequiredPlugins: map[string]*RequiredPlugin{
							"docker": {
								Name:   "docker",
								Source: "github.com/hashicorp/docker",
								Type:   &addrs.Plugin{"github.com", "hashicorp", "docker"},
								Requirement: VersionConstraint{
									Required: nil,
								},
								PluginDependencyReason: PluginDependencyImplicit,
							},
						}},
					},
				},
			}},
		{"missing-required-plugin-for-nested-post-processor", PackerConfig{
			parser: getBasicParser(func(p *Parser) {
				p.PluginConfig.PostProcessorRedirects = map[string]string{
					"docker-push": "github.com/hashicorp/docker",
				}
			},
			)},
			`
			packer {
			}`, `
			build {
				post-processors {
					post-processor "docker-push" {
					}
				}
			}
			`,
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{RequiredPlugins: map[string]*RequiredPlugin{
							"docker": {
								Name:   "docker",
								Source: "github.com/hashicorp/docker",
								Type:   &addrs.Plugin{"github.com", "hashicorp", "docker"},
								Requirement: VersionConstraint{
									Required: nil,
								},
								PluginDependencyReason: PluginDependencyImplicit,
							},
						}},
					},
				},
			}},

		{"required-plugin-renamed", PackerConfig{
			parser: getBasicParser(func(p *Parser) {
				p.PluginConfig.BuilderRedirects = map[string]string{
					"amazon-ebs": "github.com/hashicorp/amazon",
				}
			},
			)},
			`
			packer {
				required_plugins {
					amazon-v1 = {
						source  = "github.com/hashicorp/amazon"
						version = "~> v1.0"
					}
				}
			}`, `
			source "amazon-v1-ebs" "example" {
			}
			source "amazon-ebs" "example" {
			}
			`,
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{RequiredPlugins: map[string]*RequiredPlugin{
							"amazon-v1": {
								Name:   "amazon-v1",
								Source: "github.com/hashicorp/amazon",
								Type:   &addrs.Plugin{"github.com", "hashicorp", "amazon"},
								Requirement: VersionConstraint{
									Required: mustVersionConstraints(version.NewConstraint("~> v1.0")),
								},
								PluginDependencyReason: PluginDependencyExplicit,
							},
						}},
						{RequiredPlugins: map[string]*RequiredPlugin{
							"amazon": {
								Name:   "amazon",
								Source: "github.com/hashicorp/amazon",
								Type:   &addrs.Plugin{"github.com", "hashicorp", "amazon"},
								Requirement: VersionConstraint{
									Required: nil,
								},
								PluginDependencyReason: PluginDependencyImplicit,
							},
						}},
					},
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

			rest, diags := cfg.parser.ParseHCL([]byte(tt.restOfTemplate), "rest.pkr.hcl")
			if len(diags) > 0 {
				t.Fatal(diags)
			}
			if gotDiags := cfg.decodeImplicitRequiredPluginsBlocks(rest); (len(gotDiags) > 0) != tt.wantDiags {
				t.Fatal(gotDiags)
			}
			if diff := cmp.Diff(tt.wantConfig, cfg, cmpOpts...); diff != "" {
				t.Errorf("PackerConfig.inferImplicitRequiredPluginFromBlocks() unexpected PackerConfig: %v", diff)
			}
		})
	}
}
