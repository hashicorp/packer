// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package addrs

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPluginParseSourceString(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		want      *Plugin
		wantDiags bool
	}{
		{"invalid: only one component, rejected", "potato", nil, true},
		{"invalid: two components in name", "hashicorp/azr", nil, true},
		{"valid: three components, nothing superfluous", "github.com/hashicorp/azr", &Plugin{"github.com/hashicorp/azr"}, false},
		{"valid: 16 components, nothing superfluous", "github.com/hashicorp/azr/a/b/c/d/e/f/g/h/i/j/k/l/m", &Plugin{"github.com/hashicorp/azr/a/b/c/d/e/f/g/h/i/j/k/l/m"}, false},
		{"invalid: trailing slash", "github.com/hashicorp/azr/", nil, true},
		{"invalid: reject because scheme specified", "https://github.com/hashicorp/azr", nil, true},
		{"invalid: reject because query non nil", "github.com/hashicorp/azr?arg=1", nil, true},
		{"invalid: reject because fragment present", "github.com/hashicorp/azr#anchor", nil, true},
		{"invalid: leading and trailing slashes are removed", "/github.com/hashicorp/azr/", nil, true},
		{"invalid: leading slashes are removed", "/github.com/hashicorp/azr", nil, true},
		{"invalid: plugin name contains packer-", "/github.com/hashicorp/packer-azr", nil, true},
		{"invalid: plugin name contains packer-plugin-", "/github.com/hashicorp/packer-plugin-azr", nil, true},
		{"invalid: 17 components, too many parts to URL", "github.com/hashicorp/azr/a/b/c/d/e/f/g/h/i/j/k/l/m/n", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePluginSourceString(tt.source)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePluginSourceString() got = %v, want %v", got, tt.want)
			}
			if tt.wantDiags && err == nil {
				t.Errorf("Expected error, but got none")
			}
			if !tt.wantDiags && err != nil {
				t.Errorf("Unexpected error: %s", err)
			}
		})
	}
}

func TestPluginName(t *testing.T) {
	tests := []struct {
		name         string
		pluginString string
		expectName   string
	}{
		{
			"valid minimal name",
			"github.com/hashicorp/amazon",
			"amazon",
		},
		{
			// Technically we can call `Name` on a plugin created manually
			// but this is invalid as the Source's Name should not contain
			// `packer-plugin-`.
			"invalid name with prefix",
			"github.com/hashicorp/packer-plugin-amazon",
			"packer-plugin-amazon",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plug := &Plugin{
				Source: tt.pluginString,
			}

			name := plug.Name()
			if name != tt.expectName {
				t.Errorf("Expected plugin %q to have %q as name, got %q", tt.pluginString, tt.expectName, name)
			}
		})
	}
}

func TestPluginParts(t *testing.T) {
	tests := []struct {
		name          string
		pluginSource  string
		expectedParts []string
	}{
		{
			"valid with two parts",
			"factiartory.com/packer",
			[]string{"factiartory.com", "packer"},
		},
		{
			"valid with four parts",
			"factiartory.com/hashicrop/fields/packer",
			[]string{"factiartory.com", "hashicrop", "fields", "packer"},
		},
		{
			"valid, with double-slashes in the name",
			"factiartory.com/hashicrop//fields/packer//",
			[]string{"factiartory.com", "hashicrop", "fields", "packer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Plugin{tt.pluginSource}
			diff := cmp.Diff(plugin.Parts(), tt.expectedParts)
			if diff != "" {
				t.Errorf("Difference found between expected and computed parts: %s", diff)
			}
		})
	}
}
