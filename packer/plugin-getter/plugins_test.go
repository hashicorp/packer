package plugingetter

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

var (
	pluginFolderOne = filepath.Join("testdata", "plugins")

	pluginFolderTwo = filepath.Join("testdata", "plugins_2")

	pluginFolderWrongChecksums = filepath.Join("testdata", "wrong_checksums")
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
				Identifier: "github.com/hashicorp/amazon",
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
				Identifier: "github.com/hashicorp/amazon",
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
				Identifier: "github.com/hashicorp/amazon",
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
				Identifier: "github.com/hashicorp/google",
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
					pluginFolderWrongChecksums,
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
			// with the 5.0 one of an already installed plugin.
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
					pluginFolderWrongChecksums,
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
					pluginFolderWrongChecksums,
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

		{"upgrade-with-diff-protocol-version",
			// here we have something locally and test that a newer version will
			// be installed, the newer version has a lower minor protocol
			// version than the one we support.
			fields{"amazon", ">= v2"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Releases: []Release{
							{Version: "v1.2.3"},
							{Version: "v1.2.4"},
							{Version: "v1.2.5"},
							{Version: "v2.0.0"},
							{Version: "v2.1.0"},
							{Version: "v2.10.0"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"2.10.0": {{
								Filename: "packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64.zip",
								Checksum: "43156b1900dc09b026b54610c4a152edd277366a7f71ff3812583e4a35dd0d4a",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64.zip": zipFile(map[string]string{
								"packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64": "v2.10.0_x6.0_darwin_amd64",
							}),
						},
					},
				},
				[]string{
					pluginFolderWrongChecksums,
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "6", APIVersionMinor: "1",
					OS: "darwin", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},
			&Installation{
				BinaryPath: "testdata/plugins_2/github.com/hashicorp/amazon/packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64",
				Version:    "v2.10.0",
			}, false},

		{"upgrade-with-same-protocol-version",
			// here we have something locally and test that a newer version will
			// be installed.
			fields{"amazon", ">= v2"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Releases: []Release{
							{Version: "v1.2.3"},
							{Version: "v1.2.4"},
							{Version: "v1.2.5"},
							{Version: "v2.0.0"},
							{Version: "v2.1.0"},
							{Version: "v2.10.0"},
							{Version: "v2.10.1"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"2.10.1": {{
								Filename: "packer-plugin-amazon_v2.10.1_x6.1_darwin_amd64.zip",
								Checksum: "90ca5b0f13a90238b62581bbf30bacd7e2c9af6592c7f4849627bddbcb039dec",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v2.10.1_x6.1_darwin_amd64.zip": zipFile(map[string]string{
								"packer-plugin-amazon_v2.10.1_x6.1_darwin_amd64": "v2.10.1_x6.1_darwin_amd64",
							}),
						},
					},
				},
				[]string{
					pluginFolderWrongChecksums,
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "6", APIVersionMinor: "1",
					OS: "darwin", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},
			&Installation{
				BinaryPath: "testdata/plugins_2/github.com/hashicorp/amazon/packer-plugin-amazon_v2.10.1_x6.1_darwin_amd64",
				Version:    "v2.10.1",
			}, false},

		{"wrong-zip-checksum",
			// here we have something locally and test that a newer version with
			// a wrong checksum will not be installed and error.
			fields{"amazon", ">= v2"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Releases: []Release{
							{Version: "v2.10.0"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"2.10.0": {{
								Filename: "packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64.zip",
								Checksum: "133713371337133713371337c4a152edd277366a7f71ff3812583e4a35dd0d4a",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64.zip": zipFile(map[string]string{
								"packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64": "h4xx",
							}),
						},
					},
				},
				[]string{
					pluginFolderWrongChecksums,
					pluginFolderOne,
					pluginFolderTwo,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "6", APIVersionMinor: "1",
					OS: "darwin", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},

			nil, true},

		{"wrong-local-checksum",
			// here we have something wrong locally and test that a newer
			// version with a wrong checksum will not be installed
			// this should totally error.
			fields{"amazon", ">= v1"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Releases: []Release{
							{Version: "v2.10.0"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"2.10.0": {{
								Filename: "packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64.zip",
								Checksum: "133713371337133713371337c4a152edd277366a7f71ff3812583e4a35dd0d4a",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64.zip": zipFile(map[string]string{
								"packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64": "h4xx",
							}),
						},
					},
				},
				[]string{
					pluginFolderWrongChecksums,
				},
				BinaryInstallationOptions{
					APIVersionMajor: "6", APIVersionMinor: "1",
					OS: "darwin", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},

			nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Printf("starting %s test", tt.name)

			identifier, diags := addrs.ParsePluginSourceString("github.com/hashicorp/" + tt.fields.Identifier)
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
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Requirement.InstallLatest() %s", diff)
			}
			if tt.want != nil && tt.want.BinaryPath != "" {
				// Cleanup.
				// These two files should be here by now and os.Remove will fail if
				// they aren't.
				if err := os.Remove(filepath.Clean(tt.want.BinaryPath)); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(filepath.Clean(tt.want.BinaryPath + "_SHA256SUM")); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

type mockPluginGetter struct {
	Releases            []Release
	ChecksumFileEntries map[string][]ChecksumFileEntry
	Zips                map[string]io.ReadCloser
}

func (g *mockPluginGetter) Get(what string, options GetOptions) (io.ReadCloser, error) {

	var toEncode interface{}
	switch what {
	case "releases":
		toEncode = g.Releases
	case "sha256":
		toEncode = g.ChecksumFileEntries[options.version.String()]
	case "zip":
		acc := options.PluginRequirement.Identifier.Hostname + "/" +
			options.PluginRequirement.Identifier.RealRelativePath() + "/" +
			options.ExpectedZipFilename()

		zip, found := g.Zips[acc]
		if found == false {
			panic(fmt.Sprintf("could not find zipfile %s. %v", acc, g.Zips))
		}
		return zip, nil
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

func zipFile(content map[string]string) io.ReadCloser {
	buff := bytes.NewBuffer(nil)
	zipWriter := zip.NewWriter(buff)
	for fileName, content := range content {
		header := &zip.FileHeader{
			Name:             fileName,
			UncompressedSize: uint32(len([]byte(content))),
		}
		fWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(fWriter, strings.NewReader(content))
		if err != nil {
			panic(err)
		}
	}
	err := zipWriter.Close()
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(buff)
}

var _ Getter = &mockPluginGetter{}
