// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/posener/complete"
)

type InspectCommand struct {
	Meta
}

func (c *InspectCommand) Run(args []string) int {
	ctx := context.Background()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *InspectCommand) ParseArgs(args []string) (*InspectArgs, int) {
	var cfg InspectArgs
	flags := c.Meta.FlagSet("inspect")
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	cfg.AddFlagSets(flags)
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	args = flags.Args()
	if len(args) == 1 {
		cfg.Path = args[0]
	}
	return &cfg, 0
}

func (c *InspectCommand) RunContext(ctx context.Context, cla *InspectArgs) int {
	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	// here we ignore init diags to allow unknown variables to be used
	_ = packerStarter.Initialize(packer.InitializeOptions{})

	return packerStarter.InspectConfig(packer.InspectConfigOptions{
		Ui: c.Ui,
	})
}

func (*InspectCommand) Help() string {
	helpText := `
Usage: packer inspect TEMPLATE

  Inspects a template, parsing and outputting the components a template
  defines. This does not validate the contents of a template (other than
  basic syntax by necessity).

Options:

  -machine-readable  Machine-readable output
`

	return strings.TrimSpace(helpText)
}

func (c *InspectCommand) Synopsis() string {
	return "see components of a template"
}

func (c *InspectCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *InspectCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-machine-readable": complete.PredictNothing,
	}
}
