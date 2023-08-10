// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"path/filepath"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/packer"
	shell_local "github.com/hashicorp/packer/provisioner/shell-local"
	"github.com/hashicorp/packer/provisioner/sleep"
)

// testCoreConfigBuilder creates a packer CoreConfig that has a file builder
// available. This allows us to test a builder that writes files to disk.
func testCoreConfigSleepBuilder(t *testing.T) *packer.CoreConfig {
	components := packer.ComponentFinder{
		PluginConfig: &packer.PluginConfig{
			Builders: packer.MapOfBuilder{
				"file": func() (packersdk.Builder, error) { return &file.Builder{}, nil },
			},
			Provisioners: packer.MapOfProvisioner{
				"sleep":       func() (packersdk.Provisioner, error) { return &sleep.Provisioner{}, nil },
				"shell-local": func() (packersdk.Provisioner, error) { return &shell_local.Provisioner{}, nil },
			},
		},
	}
	return &packer.CoreConfig{
		Components: components,
	}
}

// testMetaFile creates a Meta object that includes a file builder
func testMetaSleepFile(t *testing.T) Meta {
	var out, err bytes.Buffer
	return Meta{
		CoreConfig: testCoreConfigSleepBuilder(t),
		Ui: &packersdk.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func TestBuildSleepTimeout(t *testing.T) {
	defer cleanup()

	c := &BuildCommand{
		Meta: testMetaSleepFile(t),
	}

	args := []string{
		filepath.Join(testFixture("timeout"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code == 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"roses.txt", "fuchsias.txt", "lilas.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}

	for _, f := range []string{"campanules.txt"} {
		if fileExists(f) {
			t.Errorf("Expected to not find %s", f)
		}
	}
}
