// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	goversion "github.com/hashicorp/go-version"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	gh "github.com/hashicorp/packer/packer/plugin-getter/github"
)

// Getter installs plugins from a remote HTTP source other than github.com,
// such as an internal mirror or an artifact repository. The host comes from
// the plugin source address and must serve the releases.hashicorp.com
// directory structure under the source's path:
//
//	packer-plugin-<name>/index.json
//	packer-plugin-<name>/<version>/packer-plugin-<name>_<version>_SHA256SUMS
//	packer-plugin-<name>/<version>/<zips named by the SHA256SUMS entries>
//
// When a version's zip names carry no plugin protocol version, its
// manifest.json supplies it, as published on releases.hashicorp.com.
// Checksum entries matching neither known naming shape are rejected, and
// nothing the remote metadata supplies is used to fetch from another origin
// or path.
type Getter struct {
	// BaseURL is the scheme and host to fetch from, e.g.
	// "https://plugins.example.com".
	BaseURL    string
	HttpClient *http.Client
	Name       string

	// index holds the versions listed by the plugin's index.json, loaded on
	// first use. A Getter instance serves a single plugin requirement, like
	// the release and github getters.
	index *pluginIndex

	// protVersions caches the protocol version of versions whose checksum
	// entries do not carry one (e.g. "1.1.4" -> "x5.0"), resolved from the
	// version's manifest.json when its checksum file is fetched.
	protVersions map[string]string
}

var _ plugingetter.Getter = &Getter{}

// protocolVersionRe matches the protocol-version field of GitHub release
// asset names, e.g. the "x5.0" in packer-plugin-comment_v0.2.12_x5.0_linux_amd64.zip.
var protocolVersionRe = regexp.MustCompile(`^x\d+\.\d+$`)

// entryHasProtocolVersion reports whether a checksum-file entry name carries
// an x-prefixed protocol version field.
var entryHasProtocolVersion = regexp.MustCompile(`_x\d+\.\d+_`)

// pluginIndex is the validated form of a plugin's index.json. Versions are
// keyed by their canonical core string (e.g. "1.1.4"); all other index
// content is ignored.
type pluginIndex struct {
	pluginType string
	versions   map[string]struct{}
}

// parseIndex validates raw index.json content into a pluginIndex.
func parseIndex(data []byte, pluginType string) (*pluginIndex, error) {
	var raw struct {
		Versions map[string]struct{} `json:"versions"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("malformed index.json: %w", err)
	}
	if len(raw.Versions) == 0 {
		return nil, fmt.Errorf("index.json contains no versions")
	}

	idx := &pluginIndex{
		pluginType: pluginType,
		versions:   map[string]struct{}{},
	}
	for k := range raw.Versions {
		ver, err := goversion.NewVersion(k)
		if err != nil {
			log.Printf("[WARN] remote-getter: ignoring unparseable version %q in index.json", k)
			continue
		}
		idx.versions[ver.String()] = struct{}{}
	}
	if len(idx.versions) == 0 {
		return nil, fmt.Errorf("index.json contains no usable versions")
	}
	return idx, nil
}

// releasesStream encodes the index versions as the json list of Release that
// Packer expects from get 'releases'.
func (idx *pluginIndex) releasesStream() (io.ReadCloser, error) {
	out := make([]plugingetter.Release, 0, len(idx.versions))
	for v := range idx.versions {
		out = append(out, plugingetter.Release{Version: "v" + v})
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(out); err != nil {
		return nil, err
	}
	return io.NopCloser(buf), nil
}

// fetchStream returns the response body of url for the caller to consume
// and close; fetch buffers it, for the small metadata files.
func (g *Getter) fetchStream(url string) (io.ReadCloser, error) {
	if g.HttpClient == nil {
		g.HttpClient = &http.Client{}
	}
	log.Printf("[DEBUG] remote-getter: getting %q", url)
	resp, err := g.HttpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	if resp.StatusCode >= 400 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s returned status %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}

func (g *Getter) fetch(url string) ([]byte, error) {
	body, err := g.fetchStream(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()
	return io.ReadAll(body)
}

// pluginPath returns the URL path of the plugin below the host, e.g.
// "mirror/hashicorp/packer-plugin-docker" for source
// "plugins.example.com/mirror/hashicorp/docker".
func pluginPath(opts plugingetter.GetOptions) (dir string, pluginType string) {
	parts := opts.PluginRequirement.Identifier.Parts()
	pluginType = "packer-plugin-" + parts[len(parts)-1]
	dir = path.Join(append(append([]string{}, parts[1:len(parts)-1]...), pluginType)...)
	return dir, pluginType
}

// loadIndex fetches and validates the plugin's index.json, reusing the
// parsed result across calls for the same plugin.
func (g *Getter) loadIndex(opts plugingetter.GetOptions) (*pluginIndex, string, error) {
	dir, pluginType := pluginPath(opts)
	if g.index != nil && g.index.pluginType == pluginType {
		return g.index, dir, nil
	}
	url := g.BaseURL + "/" + dir + "/index.json"
	data, err := g.fetch(url)
	if err != nil {
		return nil, "", err
	}
	idx, err := parseIndex(data, pluginType)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", url, err)
	}
	g.index = idx
	return idx, dir, nil
}

// versionURL resolves the version directory URL for the version being
// fetched, failing when the index does not list it.
func (g *Getter) versionURL(opts plugingetter.GetOptions) (*pluginIndex, string, error) {
	idx, dir, err := g.loadIndex(opts)
	if err != nil {
		return nil, "", err
	}
	version := opts.VersionString()
	if _, ok := idx.versions[version]; !ok {
		return nil, "", fmt.Errorf(
			"version %s of %s is not listed in %s/%s/index.json",
			version, idx.pluginType, g.BaseURL, dir)
	}
	return idx, g.BaseURL + "/" + dir + "/" + version, nil
}

// resolveProtocolVersion reads the plugin protocol version from a version's
// manifest.json and caches it for Validate.
func (g *Getter) resolveProtocolVersion(versionURL, pluginType, version string) error {
	if _, ok := g.protVersions[version]; ok {
		return nil
	}
	data, err := g.fetch(versionURL + "/" + pluginType + "_" + version + "_manifest.json")
	if err != nil {
		return fmt.Errorf(
			"%w\nThe checksum entries of this version carry no plugin protocol version in their names, so the mirrored version must include its manifest.json, as published on releases.hashicorp.com", err)
	}
	var meta plugingetter.ManifestMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return fmt.Errorf("malformed manifest.json for version %s: %w", version, err)
	}
	if meta.Metadata.ProtocolVersion == "" {
		return fmt.Errorf("manifest.json for version %s declares no protocol_version", version)
	}
	if g.protVersions == nil {
		g.protVersions = map[string]string{}
	}
	g.protVersions[version] = "x" + meta.Metadata.ProtocolVersion
	return nil
}

func (g *Getter) Get(what string, opts plugingetter.GetOptions) (io.ReadCloser, error) {
	log.Printf("[TRACE] Getting %s of %s plugin from %s", what, opts.PluginRequirement.Identifier, g.Name)

	switch what {
	case "releases":
		idx, _, err := g.loadIndex(opts)
		if err != nil {
			return nil, err
		}
		return idx.releasesStream()

	case "sha256":
		idx, versionURL, err := g.versionURL(opts)
		if err != nil {
			return nil, err
		}
		version := opts.VersionString()
		data, err := g.fetch(versionURL + "/" + idx.pluginType + "_" + version + "_SHA256SUMS")
		if err != nil {
			return nil, err
		}
		// When any listed zip carries no protocol version in its name, the
		// version's manifest.json supplies it; resolve that here, where a
		// missing manifest can fail the version loudly.
		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) != 2 || filepath.Ext(fields[1]) != ".zip" {
				continue
			}
			if !entryHasProtocolVersion.MatchString(fields[1]) {
				if err := g.resolveProtocolVersion(versionURL, idx.pluginType, version); err != nil {
					return nil, err
				}
				break
			}
		}
		return gh.TransformChecksumStream()(io.NopCloser(bytes.NewReader(data)))

	case "zip":
		_, versionURL, err := g.versionURL(opts)
		if err != nil {
			return nil, err
		}
		return g.fetchStream(versionURL + "/" + opts.ExpectedZipFilename())

	default:
		return nil, fmt.Errorf("%q not implemented", what)
	}
}

// Init parses a checksum-file entry name in either accepted shape:
//
//	packer-plugin-comment_v0.2.12_x5.0_freebsd_amd64.zip  (GitHub release asset)
//	packer-plugin-docker_1.1.4_darwin_arm64.zip           (releases.hashicorp.com)
//
// The releases.hashicorp.com shape carries no protocol version; Validate
// fills it from the version's manifest.json.
func (g *Getter) Init(req *plugingetter.Requirement, entry *plugingetter.ChecksumFileEntry) error {
	filename := entry.Filename
	res := strings.TrimPrefix(filename, req.FilenamePrefix())

	entry.Ext = filepath.Ext(res)
	res = strings.TrimSuffix(res, entry.Ext)

	parts := strings.Split(res, "_")
	switch {
	case len(parts) >= 4 && protocolVersionRe.MatchString(parts[1]):
		if !strings.HasPrefix(parts[0], "v") {
			return fmt.Errorf("malformed filename %s: version %q must have a v prefix alongside a protocol version", filename, parts[0])
		}
		entry.BinVersion, entry.ProtVersion, entry.Os, entry.Arch = parts[0], parts[1], parts[2], parts[3]
	case len(parts) == 3 && !strings.HasPrefix(parts[0], "v"):
		entry.BinVersion, entry.Os, entry.Arch = "v"+parts[0], parts[1], parts[2]
		entry.ProtVersion = ""
	default:
		return fmt.Errorf(
			"malformed filename %s: expected %s{version}_x{protocol-version}_{os}_{arch} or %s{version}_{os}_{arch}",
			filename, req.FilenamePrefix(), req.FilenamePrefix())
	}

	return nil
}

func (g *Getter) Validate(opt plugingetter.GetOptions, expectedVersion string, installOpts plugingetter.BinaryInstallationOptions, entry *plugingetter.ChecksumFileEntry) error {
	expectedBinVersion := "v" + expectedVersion
	if entry.BinVersion != expectedBinVersion {
		return fmt.Errorf("wrong version: %s does not match expected %s", entry.BinVersion, expectedBinVersion)
	}
	if entry.Os != installOpts.OS || entry.Arch != installOpts.ARCH {
		return fmt.Errorf("wrong system, expected %s_%s", installOpts.OS, installOpts.ARCH)
	}

	if entry.ProtVersion == "" {
		// The protocol version was resolved from the version's manifest.json
		// when its checksum file was fetched.
		pv, ok := g.protVersions[opt.VersionString()]
		if !ok {
			return fmt.Errorf("no protocol version known for version %s", expectedVersion)
		}
		entry.ProtVersion = pv
	}

	return installOpts.CheckProtocolVersion(entry.ProtVersion)
}

func (g *Getter) ExpectedFileName(pr *plugingetter.Requirement, version string, entry *plugingetter.ChecksumFileEntry, _ string) string {
	// The filename is rebuilt from the validated entry fields; the name the
	// remote source reported is never used, so it cannot traverse paths.
	return strings.Join([]string{
		"packer-plugin-" + pr.Identifier.Name(),
		entry.BinVersion,
		entry.ProtVersion,
		entry.Os,
		entry.Arch + entry.Ext,
	}, "_")
}
