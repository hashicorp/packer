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

- "installed" lists all installed plugins and their version. Ex: "packer plugins installed".
- "install" a plugin version with "packer plugins install <plugin> <version>".
- "required" lists required plugins. Ex: "packer plugins required <path>".
- "remove" a plugin version with "packer plugins remove <plugin> <version>".
  Omit the version parameter to remove all versions.

Related but not under the "plugins" command :

- "packer init <path>" will install all plugins required by a config.
`

	return strings.TrimSpace(helpText)
}

func (c *PluginsCommand) Run(args []string) int {
	return cli.RunResultHelp
}
