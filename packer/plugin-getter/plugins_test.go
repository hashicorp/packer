package plugingetter

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

func TestPlugin_ListInstallations(t *testing.T) {
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
					filepath.Join("testdata", "plugins"),
				},
				Extension: ".0_x4",
			},
			false,
			[]Install{
				Install{
					Version: "v1.2.3",
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Plugin.ListInstallations() = %v, want %v", got, tt.want)
			}
		})
	}
}
