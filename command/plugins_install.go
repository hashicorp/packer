// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer/hcl2template/addrs"
	"github.com/hashicorp/packer/packer"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/hashicorp/packer/packer/plugin-getter/github"
	pkrversion "github.com/hashicorp/packer/version"
	"github.com/mitchellh/cli"
)

type PluginsInstallCommand struct {
	Meta
}

func (c *PluginsInstallCommand) Synopsis() string {
	return "Install latest Packer plugin [matching version constraint]"
}

func (c *PluginsInstallCommand) Help() string {
	helpText := `
Usage: packer plugins install <plugin> [<version constraint>]

  This command will install the most recent compatible Packer plugin matching
  version constraint.
  When the version constraint is omitted, the most recent version will be
  installed.

  Ex: packer plugins install github.com/hashicorp/happycloud v1.2.3
`

	return strings.TrimSpace(helpText)
}

func (c *PluginsInstallCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	return c.RunContext(ctx, args)
}

func (c *PluginsInstallCommand) RunContext(buildCtx context.Context, args []string) int {
	if len(args) < 1 || len(args) > 2 {
		return cli.RunResultHelp
	}

	opts := plugingetter.ListInstallationsOptions{
		FromFolders: c.Meta.CoreConfig.Components.PluginConfig.KnownPluginFolders,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:              runtime.GOOS,
			ARCH:            runtime.GOARCH,
			APIVersionMajor: pluginsdk.APIVersionMajor,
			APIVersionMinor: pluginsdk.APIVersionMinor,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
		},
	}

	plugin, diags := addrs.ParsePluginSourceString(args[0])
	if diags.HasErrors() {
		c.Ui.Error(diags.Error())
		return 1
	}

	// a plugin requirement that matches them all
	pluginRequirement := plugingetter.Requirement{
		Identifier: plugin,
		Implicit:   false,
	}

	if len(args) > 1 {
		constraints, err := version.NewConstraint(args[1])
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		pluginRequirement.VersionConstraints = constraints
	}

	if runtime.GOOS == "windows" && opts.Ext == "" {
		opts.BinaryInstallationOptions.Ext = ".exe"
	}

	getters := []plugingetter.Getter{
		&github.Getter{
			// In the past some terraform plugins downloads were blocked from a
			// specific aws region by s3. Changing the user agent unblocked the
			// downloads so having one user agent per version will help mitigate
			// that a little more. Especially in the case someone forks this
			// code to make it more aggressive or something.
			// TODO: allow to set this from the config file or an environment
			// variable.
			UserAgent: "packer-getter-github-" + pkrversion.String(),
		},
	}

	newInstall, err := pluginRequirement.InstallLatest(plugingetter.InstallOptions{
		InFolders:                 opts.FromFolders,
		BinaryInstallationOptions: opts.BinaryInstallationOptions,
		Getters:                   getters,
	})

	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	if newInstall != nil {
		msg := fmt.Sprintf("Installed plugin %s %s in %q", pluginRequirement.Identifier, newInstall.Version, newInstall.BinaryPath)
		ui := &packer.ColoredUi{
			Color: packer.UiColorCyan,
			Ui:    c.Ui,
		}
		ui.Say(msg)
		return 0
	}

	return 0
}
