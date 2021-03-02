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
		// {"empty source labels", PackerConfig{parser: defaultParser}, ``, `source "" "" {}`, false, PackerConfig{}},
		{"add required_plugin", PackerConfig{parser: defaultParser}, `
		packer {
			required_plugins {
				amazon = {
					source  = "github.com/hashicorp/amazon"
					version = "~> v1.2.3"
				}
			}
		}
		`, ``, false, PackerConfig{
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
