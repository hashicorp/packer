package release

import (
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
				Filename: "packer-plugin-v0.2.12_freebsd_amd64.zip",
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
				Filename: "packer-plugin-v0.2.12.zip",
			},
			binVersion: "v0.2.12",
			wantErr:    true,
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

			if !tt.wantErr && (tt.entry.BinVersion != "0.2.12" || tt.entry.Os != "freebsd" || tt.entry.Arch != "amd64") {
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
			name: "invalid bin version",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:   "linux",
				ARCH: "amd64",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion: "1.2.3",
				Os:         "linux",
				Arch:       "amd64",
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
				BinVersion: "1.2.3",
				Os:         "darwin",
				Arch:       "amd64",
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
				BinVersion:  "1.2.3",
				Os:          "linux",
				Arch:        "arm64",
				ProtVersion: "x5.0",
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
	getter := &Getter{
		APIMajor: "5",
		APIMinor: "0",
	}
	pr := plugingetter.Requirement{
		Identifier: &addrs.Plugin{
			Source: "github.com/hashicorp/docker",
		},
	}

	entry := &plugingetter.ChecksumFileEntry{
		Os:   "linux",
		Arch: "amd64",
	}
	fileName := getter.ExpectedFileName(&pr, "1.2.3", entry, "packer-plugin-docker_v1.2.3_x5.0_linux_amd64.zip")
	assert.Equal(t, "packer-plugin-docker_v1.2.3_x5.0_linux_amd64.zip", fileName)
}
