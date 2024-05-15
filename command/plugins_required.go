// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"
	"strings"

	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
	"github.com/mitchellh/cli"
)

type PluginsRequiredCommand struct {
	Meta
}

func (c *PluginsRequiredCommand) Synopsis() string {
	return "List plugins required by a config"
}

func (c *PluginsRequiredCommand) Help() string {
	helpText := `
Usage: packer plugins required <path>

  This command will list every Packer plugin required by a Packer config, in
  packer.required_plugins blocks. All binaries matching the required version
  constrain and the current OS and Architecture will be listed. The most recent
  version (and the first of the list) will be the one picked by Packer during a
  build.

  Ex: packer plugins required require.pkr.hcl
  Ex: packer plugins required path/to/folder/
`

	return strings.TrimSpace(helpText)
}

func (c *PluginsRequiredCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *PluginsRequiredCommand) ParseArgs(args []string) (*PluginsRequiredArgs, int) {
	var cfg PluginsRequiredArgs
	flags := c.Meta.FlagSet("plugins required")
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	cfg.AddFlagSets(flags)
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	args = flags.Args()
	if len(args) != 1 {
		return &cfg, cli.RunResultHelp
	}
	cfg.Path = args[0]
	return &cfg, 0
}

func (c *PluginsRequiredCommand) RunContext(buildCtx context.Context, cla *PluginsRequiredArgs) int {

	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	// Get plugins requirements
	reqs, diags := packerStarter.PluginRequirements()
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	opts := plugingetter.ListInstallationsOptions{
		PluginDirectory: c.Meta.CoreConfig.Components.PluginConfig.PluginDirectory,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:              runtime.GOOS,
			ARCH:            runtime.GOARCH,
			Ext:             ext,
			APIVersionMajor: pluginsdk.APIVersionMajor,
			APIVersionMinor: pluginsdk.APIVersionMinor,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
		},
	}

	for _, pluginRequirement := range reqs {
		s := fmt.Sprintf("%s %s %q", pluginRequirement.Accessor, pluginRequirement.Identifier.String(), pluginRequirement.VersionConstraints.String())
		installs, err := pluginRequirement.ListInstallations(opts)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		for _, install := range installs {
			s += fmt.Sprintf(" %s", install.BinaryPath)
		}
		c.Ui.Message(s)
	}

	if len(reqs) == 0 {
		c.Ui.Message(`
No plugins requirement found, make sure you reference a Packer config
containing a packer.required_plugins block. See
https://www.packer.io/docs/templates/hcl_templates/blocks/packer
for more info.`)
	}

	return 0
}
