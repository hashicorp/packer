// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package plugingetter

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
)

var (
	pluginFolderOne = filepath.Join("testdata", "plugins")

	pluginFolderTwo = filepath.Join("testdata", "plugins_2")

	pluginFolderThree = filepath.Join("testdata", "plugins_3")

	pluginFolderWrongChecksums = filepath.Join("testdata", "wrong_checksums")
)

func TestChecksumFileEntry_init(t *testing.T) {
	expectedVersion := "v0.3.0"
	req := &Requirement{
		Identifier: &addrs.Plugin{
			Source: "github.com/ddelnano/xenserver",
		},
	}

	checkSum := &ChecksumFileEntry{
		Filename: fmt.Sprintf("packer-plugin-xenserver_%s_x5.0_darwin_amd64.zip", expectedVersion),
		Checksum: "0f5969b069b9c0a58f2d5786c422341c70dfe17bd68f896fcbd46677e8c913f1",
	}

	err := checkSum.init(req)

	if err != nil {
		t.Fatalf("ChecksumFileEntry.init failure: %v", err)
	}

	if checkSum.binVersion != expectedVersion {
		t.Errorf("failed to parse ChecksumFileEntry properly expected version '%s' but found '%s'", expectedVersion, checkSum.binVersion)
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
				pluginFolderOne,
				false,
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
				pluginFolderOne,
				false,
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
				pluginFolderOne,
				false,
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
								Checksum: "5763f8b5b5ed248894e8511a089cf399b96c7ef92d784fb30ee6242a7cb35bce",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64.zip": zipFile(map[string]string{
								// Make the false plugin echo an output that matches a subset of `describe` for install to work
								//
								// Note: this won't work on Windows as they don't have bin/sh, but this will
								// eventually be replaced by acceptance tests.
								"packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64": `#!/bin/sh
echo '{"version":"v2.10.0","api_version":"x6.0"}'`,
							}),
						},
					},
				},
				pluginFolderTwo,
				false,
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
								Checksum: "51451da5cd7f1ecd8699668d806bafe58a9222430842afbefdc62a6698dab260",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v2.10.1_x6.1_darwin_amd64.zip": zipFile(map[string]string{
								// Make the false plugin echo an output that matches a subset of `describe` for install to work
								//
								// Note: this won't work on Windows as they don't have bin/sh, but this will
								// eventually be replaced by acceptance tests.
								"packer-plugin-amazon_v2.10.1_x6.1_darwin_amd64": `#!/bin/sh
echo '{"version":"v2.10.1","api_version":"x6.1"}'`,
							}),
						},
					},
				},
				pluginFolderTwo,
				false,
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

		{"upgrade-with-one-missing-checksum-file",
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
							"2.10.0": {{
								Filename: "packer-plugin-amazon_v2.10.0_x6.1_linux_amd64.zip",
								Checksum: "5196f57f37e18bfeac10168db6915caae0341bfc4168ebc3d2b959d746cebd0a",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_v2.10.0_x6.1_linux_amd64.zip": zipFile(map[string]string{
								// Make the false plugin echo an output that matches a subset of `describe` for install to work
								//
								// Note: this won't work on Windows as they don't have bin/sh, but this will
								// eventually be replaced by acceptance tests.
								"packer-plugin-amazon_v2.10.0_x6.1_linux_amd64": `#!/bin/sh
echo '{"version":"v2.10.0","api_version":"x6.1"}'`,
							}),
						},
					},
				},
				pluginFolderTwo,
				false,
				BinaryInstallationOptions{
					APIVersionMajor: "6", APIVersionMinor: "1",
					OS: "linux", ARCH: "amd64",
					Checksummers: []Checksummer{
						{
							Type: "sha256",
							Hash: sha256.New(),
						},
					},
				},
			}},
			&Installation{
				BinaryPath: "testdata/plugins_2/github.com/hashicorp/amazon/packer-plugin-amazon_v2.10.0_x6.1_linux_amd64",
				Version:    "v2.10.0",
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
				pluginFolderTwo,
				false,
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
				pluginFolderTwo,
				false,
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
			switch tt.name {
			case "upgrade-with-diff-protocol-version",
				"upgrade-with-same-protocol-version",
				"upgrade-with-one-missing-checksum-file":
				if runtime.GOOS != "windows" {
					break
				}
				t.Skipf("Test %q cannot run on Windows because of a shell script being invoked, skipping.", tt.name)
			}

			log.Printf("starting %s test", tt.name)

			identifier, err := addrs.ParsePluginSourceString("github.com/hashicorp/" + tt.fields.Identifier)
			if err != nil {
				t.Fatalf("ParsePluginSourceString(%q): %v", tt.fields.Identifier, err)
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
		enc, ok := g.ChecksumFileEntries[options.version.String()]
		if !ok {
			return nil, fmt.Errorf("No checksum available for version %q", options.version.String())
		}
		toEncode = enc
	case "zip":
		// Note: we'll act as if the plugin sources would always be github sources for now.
		// This test will need to be updated if/when we move on to support other sources.
		parts := options.PluginRequirement.Identifier.Parts()
		acc := fmt.Sprintf("%s/%s/packer-plugin-%s/%s", parts[0], parts[1], parts[2], options.ExpectedZipFilename())

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
	return io.NopCloser(read), nil
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
	return io.NopCloser(buff)
}

var _ Getter = &mockPluginGetter{}

func Test_LessInstallList(t *testing.T) {
	tests := []struct {
		name       string
		installs   InstallList
		expectLess bool
	}{
		{
			"v1.2.1 < v1.2.2 => true",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.1",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2",
					APIVersion: "x5.0",
				},
			},
			true,
		},
		{
			// Impractical with the changes to the loading model
			"v1.2.1 = v1.2.1 => false",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.1",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.1",
					APIVersion: "x5.0",
				},
			},
			false,
		},
		{
			"v1.2.2 < v1.2.1 => false",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.1",
					APIVersion: "x5.0",
				},
			},
			false,
		},
		{
			"v1.2.2-dev < v1.2.2 => true",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2-dev",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2",
					APIVersion: "x5.0",
				},
			},
			true,
		},
		{
			"v1.2.2 < v1.2.2-dev => false",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2-dev",
					APIVersion: "x5.0",
				},
			},
			false,
		},
		{
			"v1.2.1 < v1.2.2-dev => true",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.1",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2-dev",
					APIVersion: "x5.0",
				},
			},
			true,
		},
		{
			"v1.2.3 < v1.2.2-dev => false",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.2-dev",
					APIVersion: "x5.0",
				},
			},
			false,
		},
		{
			"v1.2.3_x5.0 < v1.2.3_x5.1 => true",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.1",
				},
			},
			true,
		},
		{
			"v1.2.3_x5.0 < v1.2.3_x5.0 => false",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.0",
				},
			},
			false,
		},
		{
			"v1.2.3_x4.15 < v1.2.3_x5.0 => true",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x4.15",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.0",
				},
			},
			true,
		},
		{
			"v1.2.3_x9.0 < v1.2.3_x10.0 => true",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x9.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x10.0",
				},
			},
			true,
		},
		{
			"v1.2.3_x5.9 < v1.2.3_x5.10 => true",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.9",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.10",
				},
			},
			true,
		},
		{
			"v1.2.3_x5.0 < v1.2.3_x4.15 => false",
			InstallList{
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x5.0",
				},
				&Installation{
					BinaryPath: "host/org/plugin",
					Version:    "v1.2.3",
					APIVersion: "x4.15",
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isLess := tt.installs.Less(0, 1)
			if isLess != tt.expectLess {
				t.Errorf("Less mismatch for %s_%s < %s_%s, expected %t, got %t",
					tt.installs[0].Version,
					tt.installs[0].APIVersion,
					tt.installs[1].Version,
					tt.installs[1].APIVersion,
					tt.expectLess, isLess)
			}
		})
	}
}
