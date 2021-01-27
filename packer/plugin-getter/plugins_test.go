package plugingetter

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

var (
	pluginFolderOne = filepath.Join("testdata", "plugins")
	pluginFolderTwo = filepath.Join("testdata", "plugins_2")
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
		want    InstallList
	}{
		{
			"darwin_amazon_prot_5.0",
			fields{
				Identifier: "amazon",
			},
			ListInstallationsOptions{
				[]string{
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "5", APIVersionMinor: "0",
					OS: "darwin", ARCH: "amd64",
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
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.3_x5.0_darwin_amd64"),
				},
				{
					Version:    "v1.2.4",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.4_x5.0_darwin_amd64"),
				},
				{
					Version:    "v1.2.5",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.5_x5.0_darwin_amd64"),
				},
			},
		},
		{
			"darwin_amazon_prot_5.1",
			fields{
				Identifier: "amazon",
			},
			ListInstallationsOptions{
				[]string{
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "5", APIVersionMinor: "1",
					OS: "darwin", ARCH: "amd64",
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
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.3_x5.0_darwin_amd64"),
				},
				{
					Version:    "v1.2.4",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.4_x5.0_darwin_amd64"),
				},
				{
					Version:    "v1.2.5",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.5_x5.0_darwin_amd64"),
				},
				{
					Version:    "v1.2.6",
					BinaryPath: filepath.Join(pluginFolderTwo, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.6_x5.1_darwin_amd64"),
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
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "5", APIVersionMinor: "0",
					OS: "windows", ARCH: "amd64",
					Ext: ".exe",
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
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.3_x5.0_windows_amd64.exe"),
				},
				{
					Version:    "v1.2.4",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.4_x5.0_windows_amd64.exe"),
				},
				{
					Version:    "v1.2.5",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "amazon", "packer-plugin-amazon_v1.2.5_x5.0_windows_amd64.exe"),
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
					APIVersionMajor: "5", APIVersionMinor: "0",
					OS: "windows", ARCH: "amd64",
					Ext: ".exe",
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
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.6_x5.0_windows_amd64.exe"),
				},
				{
					Version:    "v4.5.7",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.7_x5.0_windows_amd64.exe"),
				},
				{
					Version:    "v4.5.8",
					BinaryPath: filepath.Join(pluginFolderOne, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.8_x5.0_windows_amd64.exe"),
				},
				{
					Version:    "v4.5.9",
					BinaryPath: filepath.Join(pluginFolderTwo, "github.com", "hashicorp", "google", "packer-plugin-google_v4.5.9_x5.0_windows_amd64.exe"),
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

func TestRequirement_InstallLatest(t *testing.T) {
	type fields struct {
		Identifier         string
		VersionConstraints string
	}
	type args struct {
		opts InstallOptions
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Installation
		wantErr bool
	}{
		{"already-installed-same-api-version",
			fields{"amazon", "v1.2.3"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Releases: []Release{
							{Version: "v1.2.3"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"1.2.3": {{
								// here the checksum file tells us what zipfiles
								// to expect. maybe we could cache the zip file
								// ? but then the plugin is present on the drive
								// twice.
								Filename: "packer-plugin-amazon_v1.2.3_x5.0_darwin_amd64.zip",
								Checksum: "1337c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							}},
						},
					},
				},
				[]string{
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "5", APIVersionMinor: "0",
					OS: "darwin", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},
			nil, false},

		{"already-installed-compatible-api-minor-version",
			// here 'packer' uses the procol version 5.1 which is compatible
			// with the 5.0 one.
			fields{"amazon", "v1.2.3"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Releases: []Release{
							{Version: "v1.2.3"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"1.2.3": {{
								Filename: "packer-plugin-amazon_v1.2.3_x5.0_darwin_amd64.zip",
								Checksum: "1337c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							}},
						},
					},
				},
				[]string{
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "5", APIVersionMinor: "1",
					OS: "darwin", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},
			nil, false},

		{"ignore-incompatible-higher-protocol-version",
			// here 'packer' needs a binary with protocol version 5.0, and a
			// working plugin is already installed; but a plugin with version
			// 6.0 is available locally and remotely. It simply needs to be
			// ignored.
			fields{"amazon", ">= v1"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Releases: []Release{
							{Version: "v1.2.3"},
							{Version: "v1.2.4"},
							{Version: "v1.2.5"},
							{Version: "v2.0.0"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"2.0.0": {{
								Filename: "packer-plugin-amazon_v2.0.0_x6.0_darwin_amd64.zip",
								Checksum: "1337c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							}},
							"1.2.5": {{
								Filename: "packer-plugin-amazon_v1.2.5_x5.0_darwin_amd64.zip",
								Checksum: "1337c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							}},
						},
					},
				},
				[]string{
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "5", APIVersionMinor: "0",
					OS: "darwin", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},
			nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Printf("starting %s test", tt.name)

			identifier, diags := addrs.ParsePluginSourceString(tt.fields.Identifier)
			if len(diags) != 0 {
				t.Fatalf("ParsePluginSourceString(%q): %v", tt.fields.Identifier, diags)
			}
			cts, err := version.NewConstraint(tt.fields.VersionConstraints)
			if err != nil {
				t.Fatalf("version.NewConstraint(%q): %v", tt.fields.Identifier, err)
			}
			pr := &Requirement{
				Identifier:         identifier,
				VersionConstraints: cts,
			}
			got, err := pr.InstallLatest(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Requirement.InstallLatest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Requirement.InstallLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockPluginGetter struct {
	Releases            []Release
	ChecksumFileEntries map[string][]ChecksumFileEntry
}

func (g *mockPluginGetter) Get(what string, options GetOptions) (io.ReadCloser, error) {

	var toEncode interface{}
	switch what {
	case "releases":
		toEncode = g.Releases
	case "sha256":
		toEncode = g.ChecksumFileEntries[options.version.String()]
	default:
		panic("Don't know how to get " + what)
	}

	read, write := io.Pipe()
	go func() {
		if err := json.NewEncoder(write).Encode(toEncode); err != nil {
			panic(err)
		}
	}()
	return ioutil.NopCloser(read), nil
}

var _ Getter = &mockPluginGetter{}
