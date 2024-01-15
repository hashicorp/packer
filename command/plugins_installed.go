// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"crypto/sha256"
	"log"
	"runtime"
	"strings"

	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

type PluginsInstalledCommand struct {
	Meta
}

func (c *PluginsInstalledCommand) Synopsis() string {
	return "List all installed Packer plugin binaries"
}

func (c *PluginsInstalledCommand) Help() string {
	helpText := `
Usage: packer plugins installed

  This command lists all installed plugin binaries that match with the current
  OS and architecture. Packer's API version will be ignored.

`

	return strings.TrimSpace(helpText)
}

func (c *PluginsInstalledCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	return c.RunContext(ctx)
}

func (c *PluginsInstalledCommand) RunContext(buildCtx context.Context) int {

	opts := plugingetter.ListInstallationsOptions{
		PluginDirectory: c.Meta.CoreConfig.Components.PluginConfig.PluginDirectory,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:   runtime.GOOS,
			ARCH: runtime.GOARCH,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
		},
	}

	if runtime.GOOS == "windows" && opts.Ext == "" {
		opts.BinaryInstallationOptions.Ext = ".exe"
	}

	log.Printf("[TRACE] init: %#v", opts)

	// a plugin requirement that matches them all
	allPlugins := plugingetter.Requirement{
		Accessor:           "",
		VersionConstraints: nil,
		Identifier:         nil,
	}

	installations, err := allPlugins.ListInstallations(opts)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	for _, installation := range installations {
		c.Ui.Message(installation.BinaryPath)
	}

	return 0
}
