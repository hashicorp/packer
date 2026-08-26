// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package remote

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/hcl2template/addrs"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/stretchr/testify/assert"
)

func TestParseIndex(t *testing.T) {
	golden, err := os.ReadFile("testdata/releases_hashicorp_com_index.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("parses a verbatim releases.hashicorp.com index", func(t *testing.T) {
		idx, err := parseIndex(golden, "packer-plugin-docker")
		if err != nil {
			t.Fatal(err)
		}
		if len(idx.versions) != 23 {
			t.Fatalf("expected 23 versions, got %d", len(idx.versions))
		}
		if _, ok := idx.versions["1.1.4"]; !ok {
			t.Fatal("expected version 1.1.4 in the index")
		}
	})

	t.Run("parses a minimal mirror index", func(t *testing.T) {
		idx, err := parseIndex([]byte(`{"versions": {"0.6.6": {}, "0.6.7": {}}}`), "packer-plugin-git")
		if err != nil {
			t.Fatal(err)
		}
		if len(idx.versions) != 2 {
			t.Fatalf("expected 2 versions, got %d", len(idx.versions))
		}
	})

	rejected := []struct {
		name  string
		index string
	}{
		{"no versions", `{"versions": {}}`},
		{"malformed json", `{"versions": `},
		{"only unparseable versions", `{"versions": {"not-a-version": {}}}`},
	}
	for _, tt := range rejected {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			if _, err := parseIndex([]byte(tt.index), "packer-plugin-docker"); err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}

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
			name: "github release asset shape parses",
			entry: &plugingetter.ChecksumFileEntry{
				Filename: "packer-plugin-v0.2.12_x5.0_freebsd_amd64.zip",
			},
			binVersion:      "v0.2.12",
			protocolVersion: "x5.0",
			os:              "freebsd",
			arch:            "amd64",
		},
		{
			name: "releases.hashicorp.com shape parses with the version normalized",
			entry: &plugingetter.ChecksumFileEntry{
				Filename: "packer-plugin-1.1.4_darwin_arm64.zip",
			},
			binVersion:      "v1.1.4",
			protocolVersion: "",
			os:              "darwin",
			arch:            "arm64",
		},
		{
			name: "three fields with a v prefix is neither shape",
			entry: &plugingetter.ChecksumFileEntry{
				Filename: "packer-plugin-v0.2.12_freebsd_amd64.zip",
			},
			wantErr: true,
		},
		{
			name: "four fields without a protocol version is neither shape",
			entry: &plugingetter.ChecksumFileEntry{
				Filename: "packer-plugin-v0.2.12_x5.0.zip",
			},
			wantErr: true,
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

			if !tt.wantErr && (tt.entry.BinVersion != tt.binVersion || tt.entry.ProtVersion != tt.protocolVersion || tt.entry.Os != tt.os || tt.entry.Arch != tt.arch) {
				t.Fatalf("unexpected parsed values: %+v", tt.entry)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
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
		},
		{
			name: "wrong version",
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
			name: "wrong system",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS:   "linux",
				ARCH: "amd64",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "darwin",
				Arch:        "arm64",
				ProtVersion: "x5.0",
			},
			version: "1.2.3",
			wantErr: true,
		},
		{
			name: "incompatible protocol version",
			installOpts: plugingetter.BinaryInstallationOptions{
				OS: "linux", ARCH: "amd64",
				APIVersionMajor: "5", APIVersionMinor: "0",
			},
			entry: &plugingetter.ChecksumFileEntry{
				BinVersion:  "v1.2.3",
				Os:          "linux",
				Arch:        "amd64",
				ProtVersion: "x6.0",
			},
			version: "1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getter := &Getter{}
			err := getter.Validate(plugingetter.GetOptions{}, tt.version, tt.installOpts, tt.entry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpectedFileName(t *testing.T) {
	getter := &Getter{}
	pr := &plugingetter.Requirement{
		Identifier: &addrs.Plugin{
			Source: "plugins.example.com/mirror/hashicorp/comment",
		},
	}
	entry := &plugingetter.ChecksumFileEntry{
		BinVersion:  "v1.2.3",
		ProtVersion: "x5.0",
		Os:          "linux",
		Arch:        "amd64",
		Ext:         ".zip",
	}
	fileName := getter.ExpectedFileName(pr, "1.2.3", entry, "packer-plugin-comment_v1.2.3_x5.0_linux_amd64.zip")
	assert.Equal(t, "packer-plugin-comment_v1.2.3_x5.0_linux_amd64.zip", fileName)

	// A malicious zipFileName with path traversal must not affect the output.
	maliciousFileName := getter.ExpectedFileName(pr, "1.2.3", entry,
		`packer-plugin-comment_v1.2.3_x5.0_linux_amd64_/../../tmp/PWNED.zip`)
	assert.Equal(t, "packer-plugin-comment_v1.2.3_x5.0_linux_amd64.zip", maliciousFileName)
}

// fakeRelease describes one plugin version served by the test server. All
// versions live in the releases.hashicorp.com directory structure; official
// releases carry release-site zip names and a manifest.json, retrofitted
// GitHub releases carry the author's asset names and need no manifest.
type fakeRelease struct {
	version         string   // "1.1.0"
	platforms       []string // "darwin_amd64"
	official        bool
	corruptChecksum bool
	omitManifest    bool // official only: do not serve manifest.json
	omitSums        bool // do not serve the SHA256SUMS file
}

// pluginBasePath derives the URL path the getter must request for a source,
// e.g. "example.com/mirror/hashicorp/comment" -> "/mirror/hashicorp/packer-plugin-comment".
func pluginBasePath(source string) string {
	parts := strings.Split(source, "/")
	return "/" + strings.Join(append(parts[1:len(parts)-1], "packer-plugin-"+parts[len(parts)-1]), "/")
}

// buildPluginFiles renders the served file tree for one plugin below
// pluginBasePath(source).
func buildPluginFiles(t *testing.T, source string, releases []fakeRelease) map[string][]byte {
	t.Helper()

	files := map[string][]byte{}
	base := pluginBasePath(source)
	parts := strings.Split(source, "/")
	plugin := "packer-plugin-" + parts[len(parts)-1]

	indexVersions := map[string]interface{}{}
	for _, rel := range releases {
		indexVersions[rel.version] = map[string]string{}

		sums := &strings.Builder{}
		for _, platform := range rel.platforms {
			binName := fmt.Sprintf("%s_v%s_x5.0_%s", plugin, rel.version, platform)
			zipName := binName + ".zip"
			if rel.official {
				zipName = fmt.Sprintf("%s_%s_%s.zip", plugin, rel.version, platform)
			}

			zipBuf := &bytes.Buffer{}
			zw := zip.NewWriter(zipBuf)
			fw, err := zw.Create(binName)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = fmt.Fprintf(fw, "#!/bin/sh\necho '{\"version\":\"v%s\",\"api_version\":\"x5.0\"}'\n", rel.version)
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}

			checksum := sha256.Sum256(zipBuf.Bytes())
			if rel.corruptChecksum {
				checksum[0] ^= 0xff
			}
			_, _ = fmt.Fprintf(sums, "%x  %s\n", checksum, zipName)
			files[base+"/"+rel.version+"/"+zipName] = zipBuf.Bytes()
		}

		if rel.official {
			manifestName := fmt.Sprintf("%s_%s_manifest.json", plugin, rel.version)
			manifest := []byte(`{"version": "1", "metadata": {"protocol_version": "5.0"}}`)
			checksum := sha256.Sum256(manifest)
			_, _ = fmt.Fprintf(sums, "%x  %s\n", checksum, manifestName)
			if !rel.omitManifest {
				files[base+"/"+rel.version+"/"+manifestName] = manifest
			}
		}
		if !rel.omitSums {
			sumsName := fmt.Sprintf("%s_%s_SHA256SUMS", plugin, rel.version)
			files[base+"/"+rel.version+"/"+sumsName] = []byte(sums.String())
		}
	}

	if releases != nil {
		indexBytes, err := json.Marshal(map[string]interface{}{"versions": indexVersions})
		if err != nil {
			t.Fatal(err)
		}
		files[base+"/index.json"] = indexBytes
	}

	return files
}

// newTestServer serves one or more plugin file trees, recording every
// requested path.
func newTestServer(t *testing.T, trees []map[string][]byte, requested *[]string, mu *sync.Mutex) *httptest.Server {
	t.Helper()

	files := map[string][]byte{}
	for _, tree := range trees {
		for k, v := range tree {
			files[k] = v
		}
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		*requested = append(*requested, r.URL.Path)
		mu.Unlock()
		body, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
}

// installFromServer runs InstallLatest for source against the server, into a
// fresh plugin directory.
func installFromServer(t *testing.T, server *httptest.Server, source, constraints string) (*plugingetter.Installation, error) {
	t.Helper()

	identifier, err := addrs.ParsePluginSourceString(source)
	if err != nil {
		t.Fatal(err)
	}
	cts, err := version.NewConstraint(constraints)
	if err != nil {
		t.Fatal(err)
	}

	pr := &plugingetter.Requirement{
		Identifier:         identifier,
		VersionConstraints: cts,
	}
	return pr.InstallLatest(plugingetter.InstallOptions{
		Getters: []plugingetter.Getter{
			&Getter{
				BaseURL:    server.URL,
				HttpClient: server.Client(),
				Name:       strings.Split(source, "/")[0],
			},
		},
		PluginDirectory: t.TempDir(),
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			APIVersionMajor: "5", APIVersionMinor: "0",
			OS: "darwin", ARCH: "amd64",
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
		},
	})
}

func TestInstallFromRemote(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test plugins are shell scripts, which cannot run on Windows")
	}

	tests := []struct {
		name        string
		source      string // defaults to a 4-part source
		releases    []fakeRelease
		constraints string
		wantVersion string
		wantErr     string // substring of the expected error, "" for success
	}{
		{
			name: "resolves highest version matching constraint",
			releases: []fakeRelease{
				{version: "0.9.0", platforms: []string{"darwin_amd64"}},
				{version: "1.0.5", platforms: []string{"darwin_amd64"}},
				{version: "1.1.0", platforms: []string{"darwin_amd64"}},
				{version: "2.0.0", platforms: []string{"darwin_amd64"}},
			},
			constraints: ">= 1.0, < 2.0",
			wantVersion: "v1.1.0",
		},
		{
			name: "bad checksum fails the install",
			releases: []fakeRelease{
				{version: "1.1.0", platforms: []string{"darwin_amd64"}, corruptChecksum: true},
			},
			constraints: ">= 1.0",
			wantErr:     "checksum",
		},
		{
			name: "falls back to a lower version when the platform is missing",
			releases: []fakeRelease{
				{version: "1.0.5", platforms: []string{"darwin_amd64"}},
				{version: "1.1.0", platforms: []string{"linux_arm64"}},
			},
			constraints: ">= 1.0",
			wantVersion: "v1.0.5",
		},
		{
			name:        "missing index.json fails with the URL",
			releases:    nil,
			constraints: ">= 1.0",
			wantErr:     "index.json",
		},
		{
			name:   "3-part source installs from the docroot",
			source: "ghes.example.com/ethanmdavidson/comment",
			releases: []fakeRelease{
				{version: "1.0.5", platforms: []string{"darwin_amd64"}},
			},
			constraints: ">= 1.0",
			wantVersion: "v1.0.5",
		},
		{
			name:   "5-part source installs from a nested repository path",
			source: "artifactory.example.com/artifactory/hashi-tools/hashicorp/comment",
			releases: []fakeRelease{
				{version: "1.0.5", platforms: []string{"darwin_amd64"}},
			},
			constraints: ">= 1.0",
			wantVersion: "v1.0.5",
		},
		{
			name: "verbatim releases.hashicorp.com content installs",
			releases: []fakeRelease{
				{version: "1.1.3", platforms: []string{"darwin_amd64"}, official: true},
				{version: "1.1.4", platforms: []string{"darwin_amd64"}, official: true},
			},
			constraints: ">= 1.1.0, < 1.1.4",
			wantVersion: "v1.1.3",
		},
		{
			name: "official content without its manifest fails",
			releases: []fakeRelease{
				{version: "1.1.4", platforms: []string{"darwin_amd64"}, official: true, omitManifest: true},
			},
			constraints: ">= 1.1.0",
			wantErr:     "manifest.json",
		},
		{
			name: "missing checksum file fails with its name",
			releases: []fakeRelease{
				{version: "1.1.4", platforms: []string{"darwin_amd64"}, official: true, omitSums: true},
			},
			constraints: ">= 1.1.0",
			wantErr:     "packer-plugin-comment_1.1.4_SHA256SUMS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := tt.source
			if source == "" {
				source = "example.com/mirror/hashicorp/comment"
			}
			base := pluginBasePath(source)

			var requested []string
			var mu sync.Mutex
			server := newTestServer(t, []map[string][]byte{buildPluginFiles(t, source, tt.releases)}, &requested, &mu)
			defer server.Close()

			got, err := installFromServer(t, server, source, tt.constraints)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error containing %q, got none", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error to contain %q, got: %s", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("InstallLatest() error: %s", err)
			}
			if got == nil || got.Version != tt.wantVersion {
				t.Fatalf("expected version %q installed, got %+v", tt.wantVersion, got)
			}
			if _, err := os.Stat(got.BinaryPath); err != nil {
				t.Fatalf("installed binary missing: %s", err)
			}
			if !strings.Contains(got.BinaryPath, source) {
				t.Fatalf("binary installed outside the source hierarchy: %s", got.BinaryPath)
			}

			mu.Lock()
			defer mu.Unlock()
			assert.Contains(t, requested, base+"/index.json",
				"discovery must hit the full multi-part source path")
			for _, p := range requested {
				if strings.Contains(p, "/v"+strings.TrimPrefix(tt.wantVersion, "v")+"/") {
					t.Fatalf("requested a v-prefixed version directory: %s", p)
				}
			}
		})
	}
}

// TestInstallSideBySide proves one host can serve a verbatim
// releases.hashicorp.com mirror and a retrofitted GitHub-release mirror next
// to each other.
func TestInstallSideBySide(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test plugins are shell scripts, which cannot run on Windows")
	}

	communitySource := "plugins.example.com/mirror/community/comment"
	hashicorpSource := "plugins.example.com/mirror/hashicorp/docker"

	var requested []string
	var mu sync.Mutex
	server := newTestServer(t, []map[string][]byte{
		buildPluginFiles(t, communitySource, []fakeRelease{
			{version: "0.6.7", platforms: []string{"darwin_amd64"}},
		}),
		buildPluginFiles(t, hashicorpSource, []fakeRelease{
			{version: "1.1.4", platforms: []string{"darwin_amd64"}, official: true},
		}),
	}, &requested, &mu)
	defer server.Close()

	community, err := installFromServer(t, server, communitySource, ">= 0.6.0")
	if err != nil {
		t.Fatalf("community plugin install failed: %s", err)
	}
	if community.Version != "v0.6.7" {
		t.Fatalf("expected community v0.6.7, got %+v", community)
	}

	hashicorp, err := installFromServer(t, server, hashicorpSource, ">= 1.1.0")
	if err != nil {
		t.Fatalf("hashicorp plugin install failed: %s", err)
	}
	if hashicorp.Version != "v1.1.4" {
		t.Fatalf("expected hashicorp v1.1.4, got %+v", hashicorp)
	}
}
