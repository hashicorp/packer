package hcl2template

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

func TestPackerConfig_decodeImplicitRequiredPluginsBlocks(t *testing.T) {
	type fields struct {
		PackerConfig
	}
	type args struct {
		block *hcl.Block
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantDiags  bool
		wantConfig PackerConfig
	}{
		{"invalid block", fields{PackerConfig: PackerConfig{}}, args{block: &hcl.Block{}}, false, PackerConfig{}},
		{"invalid block name", fields{PackerConfig: PackerConfig{}}, args{block: &hcl.Block{Labels: []string{""}}}, false, PackerConfig{}},
		{"implicitly require amazon plugin through datasource",
			fields{PackerConfig: PackerConfig{}},
			args{block: &hcl.Block{Labels: []string{"amazon-ami"}}},
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{
							RequiredPlugins: map[string]*RequiredPlugin{
								"amazon": {
									Name:                   "amazon",
									Source:                 "github.com/hashicorp/amazon",
									Type:                   &addrs.Plugin{"github.com", "hashicorp", "amazon"},
									PluginDependencyReason: PluginDependencyImplicit,
								},
							},
						},
					},
				},
			}},
		{"don't replace explicitly imported amazon plugin",
			fields{PackerConfig: PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{
							RequiredPlugins: map[string]*RequiredPlugin{
								"amazon": {
									Name:                   "amazon",
									Source:                 "github.com/hashicorp/amazon",
									Type:                   &addrs.Plugin{"github.com", "hashicorp", "amazon"},
									PluginDependencyReason: PluginDependencyExplicit,
								},
							},
						},
					},
				},
			}},
			args{block: &hcl.Block{Labels: []string{"amazon-ami"}}},
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{
							RequiredPlugins: map[string]*RequiredPlugin{
								"amazon": {
									Name:                   "amazon",
									Source:                 "github.com/hashicorp/amazon",
									Type:                   &addrs.Plugin{"github.com", "hashicorp", "amazon"},
									PluginDependencyReason: PluginDependencyExplicit,
								},
							},
						},
					},
				},
			}},
		{"implict import of a plugin without a dash",
			fields{PackerConfig: PackerConfig{}},
			args{block: &hcl.Block{Labels: []string{"google"}}},
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{
							RequiredPlugins: map[string]*RequiredPlugin{
								"google": {
									Name:                   "google",
									Source:                 "github.com/hashicorp/google",
									Type:                   &addrs.Plugin{"github.com", "hashicorp", "google"},
									PluginDependencyReason: PluginDependencyImplicit,
								},
							},
						},
					},
				},
			}},
		{"ignore already imported google plugin",
			fields{PackerConfig: PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{
							RequiredPlugins: map[string]*RequiredPlugin{
								"google": {
									Name:                   "google",
									Source:                 "github.com/hashicorp/google",
									Type:                   &addrs.Plugin{"github.com", "hashicorp", "google"},
									PluginDependencyReason: PluginDependencyExplicit,
								},
							},
						},
					},
				},
			}},
			args{block: &hcl.Block{Labels: []string{"google"}}},
			false,
			PackerConfig{
				Packer: struct {
					VersionConstraints []VersionConstraint
					RequiredPlugins    []*RequiredPlugins
				}{
					RequiredPlugins: []*RequiredPlugins{
						{
							RequiredPlugins: map[string]*RequiredPlugin{
								"google": {
									Name:                   "google",
									Source:                 "github.com/hashicorp/google",
									Type:                   &addrs.Plugin{"github.com", "hashicorp", "google"},
									PluginDependencyReason: PluginDependencyExplicit,
								},
							},
						},
					},
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PackerConfig{
				Packer:                  tt.fields.Packer,
				Basedir:                 tt.fields.Basedir,
				CorePackerVersionString: tt.fields.CorePackerVersionString,
				Cwd:                     tt.fields.Cwd,
				Sources:                 tt.fields.Sources,
				InputVariables:          tt.fields.InputVariables,
				LocalVariables:          tt.fields.LocalVariables,
				Datasources:             tt.fields.Datasources,
				LocalBlocks:             tt.fields.LocalBlocks,
				ValidationOptions:       tt.fields.ValidationOptions,
				Builds:                  tt.fields.Builds,
				parser:                  tt.fields.parser,
				files:                   tt.fields.files,
				except:                  tt.fields.except,
				only:                    tt.fields.only,
				force:                   tt.fields.force,
				debug:                   tt.fields.debug,
				onError:                 tt.fields.onError,
			}
			if gotDiags := cfg.inferImplicitRequiredPluginFromBlocks(tt.args.block); (len(gotDiags) > 0) != tt.wantDiags {
				t.Errorf("PackerConfig.inferImplicitRequiredPluginFromBlocks() = %v", gotDiags)
			}
			if diff := cmp.Diff(tt.wantConfig, cfg, cmpopts.IgnoreUnexported(PackerConfig{})); diff != "" {
				t.Errorf("PackerConfig.inferImplicitRequiredPluginFromBlocks() unexpected PackerConfig: %v", diff)
			}
		})
	}
}
