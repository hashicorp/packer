// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"testing"

	"github.com/hashicorp/packer/hcl2template/addrs"
	"github.com/hashicorp/packer/packer/plugin-getter/github"
	"github.com/hashicorp/packer/packer/plugin-getter/release"
	"github.com/hashicorp/packer/packer/plugin-getter/remote"
)

func TestPluginGetters(t *testing.T) {
	githubSource, err := addrs.ParsePluginSourceString("github.com/hashicorp/happycloud")
	if err != nil {
		t.Fatal(err)
	}
	getters := pluginGetters(githubSource)
	if len(getters) != 2 {
		t.Fatalf("expected the release and github getters for a github.com source, got %d getters", len(getters))
	}
	if _, ok := getters[0].(*release.Getter); !ok {
		t.Fatalf("expected the release getter first, got %T", getters[0])
	}
	if _, ok := getters[1].(*github.Getter); !ok {
		t.Fatalf("expected the github getter second, got %T", getters[1])
	}

	remoteSource, err := addrs.ParsePluginSourceString("plugins.example.com/mirror/hashicorp/happycloud")
	if err != nil {
		t.Fatal(err)
	}
	getters = pluginGetters(remoteSource)
	if len(getters) != 1 {
		t.Fatalf("expected a single remote getter for a non-github.com source, got %d getters", len(getters))
	}
	remoteGetter, ok := getters[0].(*remote.Getter)
	if !ok {
		t.Fatalf("expected a remote getter, got %T", getters[0])
	}
	if remoteGetter.BaseURL != "https://plugins.example.com" {
		t.Fatalf("wrong base URL: %s", remoteGetter.BaseURL)
	}
}
