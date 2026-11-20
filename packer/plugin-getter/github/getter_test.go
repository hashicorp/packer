// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package github

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/packer/hcl2template/addrs"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
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

func TestHostSpecificTokenAuthTransport(t *testing.T) {
	echoAuth := func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Header.Get("Authorization"))
	}
	other := httptest.NewServer(http.HandlerFunc(echoAuth))
	defer other.Close()

	var api *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/repositories/1", echoAuth)
	mux.HandleFunc("/repos/old/name", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, api.URL+"/repositories/1", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/releases/asset", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/asset", http.StatusFound)
	})
	api = httptest.NewServer(mux)
	defer api.Close()

	client := &http.Client{
		Transport: &HostSpecificTokenAuthTransport{
			TokenSources: map[string]oauth2.TokenSource{
				api.Listener.Addr().String(): oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "secret"}),
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		wantAuth string
	}{
		{
			name:     "request to the token host",
			path:     "/repositories/1",
			wantAuth: "Bearer secret",
		},
		{
			name:     "redirect within the token host",
			path:     "/repos/old/name",
			wantAuth: "Bearer secret",
		},
		{
			name:     "redirect to another host",
			path:     "/releases/asset",
			wantAuth: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(api.URL + tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			assert.Equal(t, tt.wantAuth, string(body))
		})
	}
}
