package plugingetter

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

func TestPlugin_ListInstallations(t *testing.T) {

	pluginFolderOne := filepath.Join("testdata", "plugins")

	type fields struct {
		Identifier         string
		VersionConstraints version.Constraints
	}
	tests := []struct {
		name    string
		fields  fields
		opts    ListInstallationsOptions
		wantErr bool
		want    []Install
	}{
		{
			"basic",
			fields{
				Identifier: "amazon",
			},
			ListInstallationsOptions{
				FromFolders: []string{
					pluginFolderOne,
				},
				Extension: ".0_x4",
				OS:        "darwin",
				ARCH:      "amd64",
			},
			false,
			[]Install{
				{
					Version: "v1.2.3",
					Path:    filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.3_darwin_amd64.0_x4"),
				},
			},
		},
		{
			"basic",
			fields{
				Identifier: "amazon",
			},
			ListInstallationsOptions{
				FromFolders: []string{
					pluginFolderOne,
				},
				Extension: ".0_x4.exe",
				OS:        "windows",
				ARCH:      "amd64",
			},
			false,
			[]Install{
				{
					Version: "v1.2.3",
					Path:    filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.3_windows_amd64.0_x4.exe"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identifier, diags := addrs.ParsePluginSourceString(tt.fields.Identifier)
			if diags.HasErrors() {
				t.Fatalf("%v", diags)
			}
			p := Plugin{
				Identifier:         identifier,
				VersionConstraints: tt.fields.VersionConstraints,
			}
			got, err := p.ListInstallations(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Plugin.ListInstallations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Plugin.ListInstallations() unexpected output: %s", diff)
			}
		})
	}
}
