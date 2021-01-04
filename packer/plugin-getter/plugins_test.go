package plugingetter

import (
	"crypto/sha256"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

func TestPlugin_ListInstallations(t *testing.T) {

	pluginFolderOne := filepath.Join("testdata", "plugins")
	pluginFolderTwo := filepath.Join("testdata", "plugins_2")

	type fields struct {
		Identifier         string
		VersionConstraints version.Constraints
	}
	tests := []struct {
		name    string
		fields  fields
		opts    ListInstallationsOptions
		wantErr bool
		want    InstallList
	}{
		{
			"darwin_amazon",
			fields{
				Identifier: "amazon",
			},
			ListInstallationsOptions{
				[]string{
					pluginFolderOne,
				},
				BinaryInstallationOptions{
					Extension: "_x4",
					OS:        "darwin",
					ARCH:      "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			},
			false,
			[]*Installation{
				{
					Version:    "v1.2.3",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.3_darwin_amd64.0_x4"),
				},
				{
					Version:    "v1.2.4",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.4_darwin_amd64.0_x4"),
				},
				{
					Version:    "v1.2.5",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.5_darwin_amd64.0_x4"),
				},
			},
		},
		{
			"windows_amazon",
			fields{
				Identifier: "amazon",
			},
			ListInstallationsOptions{
				[]string{
					pluginFolderOne,
				},
				BinaryInstallationOptions{
					Extension: "_x4.exe",
					OS:        "windows",
					ARCH:      "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			},
			false,
			[]*Installation{
				{
					Version:    "v1.2.3",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.3_windows_amd64.0_x4.exe"),
				},
				{
					Version:    "v1.2.4",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.4_windows_amd64.0_x4.exe"),
				},
				{
					Version:    "v1.2.5",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.5_windows_amd64.0_x4.exe"),
				},
			},
		},
		{
			"windows_google_multifolder",
			fields{
				Identifier: "hashicorp/google",
			},
			ListInstallationsOptions{
				[]string{
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					Extension: "_x4.exe",
					OS:        "windows",
					ARCH:      "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			},
			false,
			[]*Installation{
				{
					Version:    "v4.5.6",
					BinaryPath: filepath.Join(pluginFolderTwo, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.6_windows_amd64.0_x4.exe"),
				},
				{
					Version:    "v4.5.7",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.7_windows_amd64.0_x4.exe"),
				},
				{
					Version:    "v4.5.8",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.8_windows_amd64.0_x4.exe"),
				},
				{
					Version:    "v4.5.9",
					BinaryPath: filepath.Join(pluginFolderTwo, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.9_windows_amd64.0_x4.exe"),
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
			p := Requirement{
				Identifier:         identifier,
				VersionConstraints: tt.fields.VersionConstraints,
			}
			got, err := p.ListInstallations(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Plugin.ListInstallations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Plugin.ListInstallations() unexpected output: %s", diff)
			}
		})
	}
}
