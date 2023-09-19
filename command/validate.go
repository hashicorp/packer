// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"strings"

	"github.com/posener/complete"
)

type ValidateCommand struct {
	Meta
}

func (c *ValidateCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *ValidateCommand) ParseArgs(args []string) (*ValidateArgs, int) {
	var cfg ValidateArgs

	flags := c.Meta.FlagSet("validate", FlagSetBuildFilter|FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	cfg.AddFlagSets(flags)
	if err := flags.Parse(args); err != nil {
		return &cfg, 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return &cfg, 1
	}
	cfg.Path = args[0]
	return &cfg, 0
}

func (c *ValidateCommand) RunContext(ctx context.Context, cla *ValidateArgs) int {
	// By default we want to inform users of undeclared variables when validating but not during build time.
	cla.MetaArgs.WarnOnUndeclaredVar = true
	if cla.NoWarnUndeclaredVar {
		cla.MetaArgs.WarnOnUndeclaredVar = false
	}

	cfg, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return 1
	}

	diags := cfg.DetectPluginBinaries()
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	sched := NewScheduler(cfg, c.Ui, ctx)
	ret = sched.Validate(cla)

	if ret == 0 {
		c.Ui.Say("The configuration is valid.")
	}

	return ret
}

func (*ValidateCommand) Help() string {
	helpText := `
Usage: packer validate [options] TEMPLATE

  Checks the template is valid by parsing the template and also
  checking the configuration with the various builders, provisioners, etc.

  If it is not valid, the errors will be shown and the command will exit
  with a non-zero exit status. If it is valid, it will exit with a zero
  exit status.

Options:

  -syntax-only                  Only check syntax. Do not verify config of the template.
  -except=foo,bar,baz           Validate all builds other than these.
  -only=foo,bar,baz             Validate only these builds.
  -machine-readable             Produce machine-readable output.
  -var 'key=value'              Variable for templates, can be used multiple times.
  -var-file=path                JSON or HCL2 file containing user variables, can be used multiple times.
  -no-warn-undeclared-var       Disable warnings for user variable files containing undeclared variables.
  -evaluate-datasources         Evaluate data sources during validation (HCL2 only, may incur costs); Defaults to false. 
`

	return strings.TrimSpace(helpText)
}

func (*ValidateCommand) Synopsis() string {
	return "check that a template is valid"
}

func (*ValidateCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*ValidateCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-syntax-only":      complete.PredictNothing,
		"-except":           complete.PredictNothing,
		"-only":             complete.PredictNothing,
		"-var":              complete.PredictNothing,
		"-machine-readable": complete.PredictNothing,
		"-var-file":         complete.PredictNothing,
	}
}
