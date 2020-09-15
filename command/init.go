package command

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/version"
)

// InitCommand initializes a Packer working directory.
type InitCommand struct {
	Meta
}

func (c *InitCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cla, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cla)
}

func (c *InitCommand) ParseArgs(args []string) (*InitArgs, int) {
	var cfg InitArgs
	flags := c.Meta.FlagSet("init", FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	cfg.AddFlagSets(flags)
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	return &cfg, 0
}

func (c *InitCommand) RunContext(ctx context.Context, cla *InitArgs) int {
	fmt.Printf(`%s

Packer initialized with no template!

The directory has no Packer templates. You may begin working
with Packer immediately by creating a Packer template.
`, version.FormattedVersion())

	return 0
}

func (c *InitCommand) Help() string {
	helpText := `
Usage: packer init [options]

  Initialize a new or existing Packer working directory by downloading
  builders, provisioners, and post-processors defined in the template.

  This is the first command that should be executed when working with a new
  or existing template. Running this command in an empty directory will
  will perform no operation, and will need to be executed once a template
  has been created to initialize the working directory.

  It is safe to run init multiple times on a template to update the builders,
  provisioners, post-processors with changes in the template configuration file.

Options:
  -get-plugins=false                Skips plugin installation.
  -plugin-dir=PATH                  Skips plugin installation and loads plugins only from the specified directory.
`
	return helpText

}

func (c *InitCommand) Synopsis() string {
	return "Initializes a Packer working directory"
}
