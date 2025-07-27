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
)

func TestRequirement_InstallLatestFromGithub(t *testing.T) {
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
						Name: "github.com",
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
						Name: "github.com",
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
						Name: "github.com",
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
						Name: "github.com",
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
						Name: "github.com",
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
						Name: "github.com",
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
						Name: "github.com",
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
						Name: "github.com",
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

func TestRequirement_InstallLatestFromOfficialRelease(t *testing.T) {
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
						Name: "releases.hashicorp.com",
						Releases: []Release{
							{Version: "v1.2.3"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"1.2.3": {{
								// here the checksum file tells us what zipfiles
								// to expect. maybe we could cache the zip file
								// ? but then the plugin is present on the drive
								// twice.
								Filename: "packer-plugin-amazon_1.2.3_darwin_amd64.zip",
								Checksum: "1337c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							}},
						},
						Manifest: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_1.2.3_darwin_amd64_manifest.json": manifestFile(map[string]map[string]string{
								"metadata": {"protocol_version": "5.0"},
							}),
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
						Name: "releases.hashicorp.com",
						Releases: []Release{
							{Version: "v1.2.3"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"1.2.3": {{
								Filename: "packer-plugin-amazon_1.2.3_darwin_amd64.zip",
								Checksum: "1337c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							}},
						},
						Manifest: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_1.2.3_darwin_amd64_manifest.json": manifestFile(map[string]map[string]string{
								"metadata": {"protocol_version": "5.0"},
							}),
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
						Name: "releases.hashicorp.com",
						Releases: []Release{
							{Version: "v1.2.3"},
							{Version: "v1.2.4"},
							{Version: "v1.2.5"},
							{Version: "v2.0.0"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"1.2.5": {{
								Filename: "packer-plugin-amazon_1.2.5_darwin_amd64.zip",
								Checksum: "1337c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
							}},
						},
						Manifest: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_1.2.5_darwin_amd64_manifest.json": manifestFile(map[string]map[string]string{
								"metadata": {"protocol_version": "5.0"},
							}),
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
						Name: "releases.hashicorp.com",
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
								Filename: "packer-plugin-amazon_2.10.0_darwin_amd64.zip",
								Checksum: "5763f8b5b5ed248894e8511a089cf399b96c7ef92d784fb30ee6242a7cb35bce",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_2.10.0_darwin_amd64.zip": zipFile(map[string]string{
								// Make the false plugin echo an output that matches a subset of `describe` for install to work
								//
								// Note: this won't work on Windows as they don't have bin/sh, but this will
								// eventually be replaced by acceptance tests.
								"packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64": `#!/bin/sh
echo '{"version":"v2.10.0","api_version":"x6.0"}'`,
							}),
						},
						Manifest: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_2.10.0_darwin_amd64_manifest.json": manifestFile(map[string]map[string]string{
								"metadata": {"protocol_version": "6.0"},
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

		{"wrong-zip-checksum",
			// here we have something locally and test that a newer version with
			// a wrong checksum will not be installed and error.
			fields{"amazon", ">= v2"},
			args{InstallOptions{
				[]Getter{
					&mockPluginGetter{
						Name: "releases.hashicorp.com",
						Releases: []Release{
							{Version: "v2.10.0"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"2.10.0": {{
								Filename: "packer-plugin-amazon_2.10.0_darwin_amd64.zip",
								Checksum: "133713371337133713371337c4a152edd277366a7f71ff3812583e4a35dd0d4a",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_2.10.0_darwin_amd64.zip": zipFile(map[string]string{
								"packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64": "h4xx",
							}),
						},
						Manifest: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_2.10.0_darwin_amd64_manifest.json": manifestFile(map[string]map[string]string{
								"metadata": {"protocol_version": "6.0"},
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
						Name: "releases.hashicorp.com",
						Releases: []Release{
							{Version: "v2.10.0"},
						},
						ChecksumFileEntries: map[string][]ChecksumFileEntry{
							"2.10.0": {{
								Filename: "packer-plugin-amazon_2.10.0_darwin_amd64.zip",
								Checksum: "133713371337133713371337c4a152edd277366a7f71ff3812583e4a35dd0d4a",
							}},
						},
						Zips: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_2.10.0_darwin_amd64.zip": zipFile(map[string]string{
								"packer-plugin-amazon_v2.10.0_x6.0_darwin_amd64": "h4xx",
							}),
						},
						Manifest: map[string]io.ReadCloser{
							"github.com/hashicorp/packer-plugin-amazon/packer-plugin-amazon_2.10.0_darwin_amd64_manifest.json": manifestFile(map[string]map[string]string{
								"metadata": {"protocol_version": "6.0"},
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
	Name                string
	APIMajor            string
	APIMinor            string
	Manifest            map[string]io.ReadCloser
}

func (g *mockPluginGetter) Init(req *Requirement, entry *ChecksumFileEntry) error {
	filename := entry.Filename
	res := strings.TrimPrefix(filename, req.FilenamePrefix())
	// res now looks like v0.2.12_x5.0_freebsd_amd64.zip

	entry.Ext = filepath.Ext(res)

	res = strings.TrimSuffix(res, entry.Ext)
	// res now looks like v0.2.12_x5.0_freebsd_amd64

	parts := strings.Split(res, "_")
	// ["v0.2.12", "x5.0", "freebsd", "amd64"]

	if g.Name == "github.com" {
		if len(parts) < 4 {
			return fmt.Errorf("malformed filename expected %s{version}_x{protocol-version}_{os}_{arch}", req.FilenamePrefix())
		}

		entry.BinVersion, entry.ProtVersion, entry.Os, entry.Arch = parts[0], parts[1], parts[2], parts[3]
	} else {
		if len(parts) < 3 {
			return fmt.Errorf("malformed filename expected %s{version}_{os}_{arch}", req.FilenamePrefix())
		}

		entry.BinVersion, entry.Os, entry.Arch = parts[0], parts[1], parts[2]
		entry.BinVersion = strings.TrimPrefix(entry.BinVersion, "v")
	}

	return nil
}

func (g *mockPluginGetter) Validate(opt GetOptions, expectedVersion string, installOpts BinaryInstallationOptions, entry *ChecksumFileEntry) error {
	if g.Name == "github.com" {
		expectedBinVersion := "v" + expectedVersion

		if entry.BinVersion != expectedBinVersion {
			return fmt.Errorf("wrong version: %s does not match expected %s", entry.BinVersion, expectedBinVersion)
		}
		if entry.Os != installOpts.OS || entry.Arch != installOpts.ARCH {
			return fmt.Errorf("wrong system, expected %s_%s", installOpts.OS, installOpts.ARCH)
		}

		return installOpts.CheckProtocolVersion(entry.ProtVersion)
	} else {
		if entry.BinVersion != expectedVersion {
			return fmt.Errorf("wrong version: %s does not match expected %s", entry.BinVersion, expectedVersion)
		}
		if entry.Os != installOpts.OS || entry.Arch != installOpts.ARCH {
			return fmt.Errorf("wrong system, expected %s_%s got %s_%s", installOpts.OS, installOpts.ARCH, entry.Os, entry.Arch)
		}

		manifest, err := g.Get("meta", opt)
		if err != nil {
			return err
		}

		var data ManifestMeta
		body, err := io.ReadAll(manifest)
		if err != nil {
			log.Printf("Failed to unmarshal manifest json: %s", err)
			return err
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Printf("Failed to unmarshal manifest json: %s", err)
			return err
		}

		err = installOpts.CheckProtocolVersion("x" + data.Metadata.ProtocolVersion)
		if err != nil {
			return err
		}

		g.APIMajor = strings.Split(data.Metadata.ProtocolVersion, ".")[0]
		g.APIMinor = strings.Split(data.Metadata.ProtocolVersion, ".")[1]

		log.Printf("#### versions API %s.%s, entry %s.%s", g.APIMajor, g.APIMinor, entry.ProtVersion, entry.BinVersion)

		return nil
	}
}

func (g *mockPluginGetter) ExpectedFileName(pr *Requirement, version string, entry *ChecksumFileEntry, zipFileName string) string {
	if g.Name == "github.com" {
		return zipFileName
	} else {
		pluginSourceParts := strings.Split(pr.Identifier.Source, "/")

		// We need to verify that the plugin source is in the expected format
		return strings.Join([]string{fmt.Sprintf("packer-plugin-%s", pluginSourceParts[2]),
			"v" + version,
			"x" + g.APIMajor + "." + g.APIMinor,
			entry.Os,
			entry.Arch + ".zip",
		}, "_")
	}
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
	case "meta":
		// Note: we'll act as if the plugin sources would always be github sources for now.
		// This test will need to be updated if/when we move on to support other sources.
		parts := options.PluginRequirement.Identifier.Parts()
		acc := fmt.Sprintf("%s/%s/packer-plugin-%s/packer-plugin-%s_%s_%s_%s_manifest.json", parts[0], parts[1], parts[2], parts[2], options.version, options.BinaryInstallationOptions.OS, options.BinaryInstallationOptions.ARCH)

		manifest, found := g.Manifest[acc]
		if found == false {
			panic(fmt.Sprintf("could not find manifest file %s. %v", acc, g.Zips))
		}
		return manifest, nil
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

func manifestFile(content map[string]map[string]string) io.ReadCloser {
	jsonBytes, err := json.Marshal(content)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}

	buffer := bytes.NewBuffer(jsonBytes)
	return io.NopCloser(buffer)
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
