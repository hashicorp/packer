// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	gh "github.com/hashicorp/packer/packer/plugin-getter/github"
)

type Getter struct {
	HttpClient *http.Client
	Name       string
}

var _ plugingetter.Getter = &Getter{}

func transformZipStream() func(in io.ReadCloser) (io.ReadCloser, error) {
	return func(in io.ReadCloser) (io.ReadCloser, error) {
		defer in.Close()
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, in)
		if err != nil {
			panic(err)
		}
		return io.NopCloser(buf), nil
	}
}

// transformReleasesVersionStream get a stream from github tags and transforms it into
// something Packer wants, namely a json list of Release.
func transformReleasesVersionStream(in io.ReadCloser) (io.ReadCloser, error) {
	if in == nil {
		return nil, fmt.Errorf("transformReleasesVersionStream got nil body")
	}
	defer in.Close()
	dec := json.NewDecoder(in)

	var m gh.PluginMetadata
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	var out []plugingetter.Release
	for _, m := range m.Versions {
		out = append(out, plugingetter.Release{
			Version: "v" + m.Version,
		})
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(out); err != nil {
		return nil, err
	}

	return io.NopCloser(buf), nil
}

func (g *Getter) Get(what string, opts plugingetter.GetOptions) (io.ReadCloser, error) {
	log.Printf("[TRACE] Getting %s of %s plugin from %s", what, opts.PluginRequirement.Identifier, g.Name)
	// The gitHub plugin we are using because we are not changing the plugin source string, if we decide to change that,
	// then we need to write this method for release getter as well, but that will change the packer init and install command as well
	ghURI, err := gh.NewGithubPlugin(opts.PluginRequirement.Identifier)
	if err != nil {
		return nil, err
	}

	if g.HttpClient == nil {
		g.HttpClient = &http.Client{}
	}

	var req *http.Request
	transform := transformZipStream()

	switch what {
	case "releases":
		url := filepath.ToSlash("https://releases.hashicorp.com/" + ghURI.PluginType() + "/index.json")
		req, err = http.NewRequest("GET", url, nil)
		transform = transformReleasesVersionStream
	case "sha256":
		// https://github.com/sylviamoss/packer-plugin-comment/releases/download/v0.2.11/packer-plugin-comment_v0.2.11_x5_SHA256SUMS
		url := filepath.ToSlash("https://releases.hashicorp.com/" + ghURI.PluginType() + "/" + opts.VersionString() + "/" + ghURI.PluginType() + "_" + opts.VersionString() + "_SHA256SUMS")
		transform = gh.TransformChecksumStream()
		req, err = http.NewRequest("GET", url, nil)
	case "zip":
		// https://releases.hashicorp.com/terraform-provider-akamai/8.0.0/terraform-provider-akamai_8.0.0_darwin_arm64.zip
		url := filepath.ToSlash("https://releases.hashicorp.com/" + ghURI.PluginType() + "/" + opts.VersionString() + "/" + opts.ExpectedZipFilename())
		req, err = http.NewRequest("GET", url, nil)
	default:
		return nil, fmt.Errorf("%q not implemented", what)
	}

	if err != nil {
		log.Printf("[ERROR] http-getter: error creating request for %q: %s", what, err)
		return nil, err
	}

	resp, err := g.HttpClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		log.Printf("[ERROR] Got error while trying getting data from releases.hashicorp.com, %v", err)
		return nil, plugingetter.HTTPFailure
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Printf("[ERROR] http-getter: error closing response body: %s", err)
		}
	}(resp.Body)

	return transform(resp.Body)
}

// Init method : a file inside will look like so:
//
//	packer-plugin-comment_0.2.12_freebsd_amd64.zip
func (g *Getter) Init(req *plugingetter.Requirement, entry *plugingetter.ChecksumFileEntry) error {
	filename := entry.Filename
	//remove the test line below where hardcoded prefix being used
	res := strings.TrimPrefix(filename, req.FilenamePrefix())
	// res now looks like v0.2.12_freebsd_amd64.zip

	entry.Ext = filepath.Ext(res)

	res = strings.TrimSuffix(res, entry.Ext)
	// res now looks like 0.2.12_freebsd_amd64

	parts := strings.Split(res, "_")
	// ["0.2.12", "freebsd", "amd64"]
	if len(parts) < 3 {
		return fmt.Errorf("malformed filename expected %s{version}_{os}_{arch}", req.FilenamePrefix())
	}

	entry.BinVersion, entry.Os, entry.Arch = parts[0], parts[1], parts[2]
	entry.BinVersion = strings.TrimPrefix(entry.BinVersion, "v")

	return nil
}

func (g *Getter) Validate(expectedVersion string, installOpts plugingetter.BinaryInstallationOptions, entry *plugingetter.ChecksumFileEntry) error {
	if entry.BinVersion != expectedVersion {
		return fmt.Errorf("wrong version: '%s' does not match expected %s ", entry.BinVersion, expectedVersion)
	}
	if entry.Os != installOpts.OS || entry.Arch != installOpts.ARCH {
		return fmt.Errorf("wrong system, expected %s_%s got %s_%s", installOpts.OS, installOpts.ARCH, entry.Os, entry.Arch)
	}

	return nil
}

func (g *Getter) ExpectedFileName(pr *plugingetter.Requirement, version string, entry *plugingetter.ChecksumFileEntry, zipFileName string) string {
	pluginSourceParts := strings.Split(pr.Identifier.Source, "/")

	// We need to verify that the plugin source is in the expected format
	return strings.Join([]string{fmt.Sprintf("packer-plugin-%s", pluginSourceParts[2]),
		"v" + version,
		"x" + pluginsdk.APIVersionMajor + "." + pluginsdk.APIVersionMinor,
		entry.Os,
		entry.Arch + ".zip",
	}, "_")
}
