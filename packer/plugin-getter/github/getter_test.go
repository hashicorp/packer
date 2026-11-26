// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package github

import (
	"io"
	"strings"
	"testing"

	"github.com/hashicorp/packer/hcl2template/addrs"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {

	tests := []struct {
		name            string
		entry           *plugingetter.ChecksumFileEntry
		binVersion      string
		protocolVersion string
		os              string
		arch            string
		wantErr         bool
	}{
		{
			name: "valid format parses",
			entry: &plugingetter.ChecksumFileEntry{
				Filename: "packer-plugin-v0.2.12_x5.0_freebsd_amd64.zip",
			},
			binVersion:      "v0.2.12",
			protocolVersion: "x5.0",
			os:              "freebsd",
			arch:            "amd64",
			wantErr:         false,
		},
		{
			name: "malformed filename returns error",
			entry: &plugingetter.ChecksumFileEntry{
				Filename: "packer-plugin-v0.2.12_x5.0.zip",
			},
			binVersion:      "v0.2.12",
			protocolVersion: "x5.0",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &plugingetter.Requirement{}

			getter := &Getter{}
			err := getter.Init(req, tt.entry)

			if err != nil && !tt.wantErr {
				t.Fatalf("unexpected error: %s", err)
			}

			if err == nil && tt.wantErr {
				t.Fatal("expected error but got nil")
			}

			if !tt.wantErr && (tt.entry.BinVersion != "v0.2.12" || tt.entry.ProtVersion != "x5.0" || tt.entry.Os != "freebsd" || tt.entry.Arch != "amd64") {
				t.Fatalf("unexpected parsed values: %+v", tt.entry)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		option      plugingetter.GetOptions
		installOpts plugingetter.BinaryInstallationOptions
		entry       *plugingetter.ChecksumFileEntry
		version     string
		wantErr     bool
	}{
		{
			name: "valid entry",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:   "linux",
				ARCH: "amd64",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "linux",
				Arch:        "amd64",
				ProtVersion: "x5.0",
			},
			version: "1.2.3",
			wantErr: false,
		},
		{
			name: "invalid bin version",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:   "linux",
				ARCH: "amd64",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "linux",
				Arch:        "amd64",
				ProtVersion: "x5.0",
			},
			version: "1.2.4",
			wantErr: true,
		},
		{
			name: "wrong OS",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:   "linux",
				ARCH: "amd64",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "darwin",
				Arch:        "amd64",
				ProtVersion: "x5.0",
			},
			version: "1.2.3",
			wantErr: true,
		},
		{
			name: "wrong Arch",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:   "linux",
				ARCH: "amd64",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "linux",
				Arch:        "arm64",
				ProtVersion: "x5.0",
			},
			version: "1.2.3",
			wantErr: true,
		},
		{
			name: "invalid API major version",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:              "linux",
				ARCH:            "amd64",
				APIVersionMajor: "5",
				APIVersionMinor: "0",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "linux",
				Arch:        "amd64",
				ProtVersion: "x4.0",
			},
			version: "1.2.3",
			wantErr: true,
		},
		{
			name: "invalid API minor version",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:              "linux",
				ARCH:            "amd64",
				APIVersionMajor: "5",
				APIVersionMinor: "0",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "linux",
				Arch:        "amd64",
				ProtVersion: "x5.4",
			},
			version: "1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			getter := &Getter{}
			err := getter.Validate(plugingetter.GetOptions{}, tt.version, tt.installOpts, tt.entry)

			if err != nil && !tt.wantErr {
				t.Fatalf("unexpected error: %s", err)
			}

			if err == nil && tt.wantErr {
				t.Fatal("expected error but got nil")
			}
		})
	}
}

func TestExpectedFileName(t *testing.T) {
	getter := &Getter{}
	pr := plugingetter.Requirement{
		Identifier: &addrs.Plugin{
			Source: "github.com/hashicorp/docker",
		},
	}
	entry := &plugingetter.ChecksumFileEntry{
		BinVersion:  "v1.2.3",
		ProtVersion: "x5.0",
		Os:          "linux",
		Arch:        "amd64",
		Ext:         ".zip",
	}
	// The raw zipFileName argument must be ignored; the result is built from
	// the validated entry fields, preventing path traversal via a crafted filename.
	fileName := getter.ExpectedFileName(&pr, "1.2.3", entry, "packer-plugin-docker_v1.2.3_x5.0_linux_amd64.zip")
	assert.Equal(t, "packer-plugin-docker_v1.2.3_x5.0_linux_amd64.zip", fileName)

	// A malicious zipFileName with path traversal must not affect the output.
	maliciousFileName := getter.ExpectedFileName(&pr, "1.2.3", entry,
		`packer-plugin-docker_v1.2.3_x5.0_linux_amd64_/../../tmp/PWNED.zip`)
	assert.Equal(t, "packer-plugin-docker_v1.2.3_x5.0_linux_amd64.zip", maliciousFileName)
}

func TestTransformChecksumStream(t *testing.T) {
	entryFoo := plugingetter.ChecksumFileEntry{
		Checksum: "1d6a2e1a7b2f3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d",
		Filename: "packer-plugin-foo_v1.0.0_x5.0_linux_amd64.zip",
	}
	entryBar := plugingetter.ChecksumFileEntry{
		Checksum: "2d6a2e1a7b2f3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d",
		Filename: "packer-plugin-bar_v1.0.0_x5.0_linux_amd64.zip",
	}

	line := func(e plugingetter.ChecksumFileEntry) string {
		return e.Checksum + "  " + e.Filename + "\n"
	}

	tests := []struct {
		name  string
		input string
		want  []plugingetter.ChecksumFileEntry
	}{
		{
			name:  "nominal file parses",
			input: line(entryFoo) + line(entryBar),
			want:  []plugingetter.ChecksumFileEntry{entryFoo, entryBar},
		},
		{
			name:  "leading blank line is skipped",
			input: "\n" + line(entryFoo) + line(entryBar),
			want:  []plugingetter.ChecksumFileEntry{entryFoo, entryBar},
		},
		{
			name:  "blank line between entries is skipped",
			input: line(entryFoo) + "\n" + line(entryBar),
			want:  []plugingetter.ChecksumFileEntry{entryFoo, entryBar},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := TransformChecksumStream()(io.NopCloser(strings.NewReader(tt.input)))
			if err != nil {
				t.Fatalf("TransformChecksumStream returned an error: %s", err)
			}

			entries, err := plugingetter.ParseChecksumFileEntries(out)
			if err != nil {
				t.Fatalf("could not parse transformed checksum stream as JSON: %s", err)
			}

			assert.Equal(t, tt.want, entries)
		})
	}
}
