// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"strings"

	"github.com/mitchellh/cli"
)

type PluginsCommand struct {
	Meta
}

func (c *PluginsCommand) Synopsis() string {
	return "Interact with Packer plugins and catalog"
}

func (c *PluginsCommand) Help() string {
	helpText := `
Usage: packer plugins <subcommand> [options] [args]
  This command groups subcommands for interacting with Packer plugins.

Related but not under the "plugins" command :

- "packer init <path>" will install all plugins required by a config.
`

	return strings.TrimSpace(helpText)
}

func (c *PluginsCommand) Run(args []string) int {
	return cli.RunResultHelp
}
