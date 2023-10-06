// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/hashicorp/packer/helper/wrappedreadline"
	"github.com/hashicorp/packer/helper/wrappedstreams"
	"github.com/hashicorp/packer/packer"
	"github.com/posener/complete"
)

var TiniestBuilder = strings.NewReader(`{
	"builders": [
		{
			"type":"null",
			"communicator": "none"
		}
	]
}`)

type ConsoleCommand struct {
	Meta
}

func (c *ConsoleCommand) Run(args []string) int {
	ctx := context.Background()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *ConsoleCommand) ParseArgs(args []string) (*ConsoleArgs, int) {
	var cfg ConsoleArgs
	flags := c.Meta.FlagSet("console")
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

func (c *ConsoleCommand) RunContext(ctx context.Context, cla *ConsoleArgs) int {
	packerStarter, ret := c.GetConfig(&cla.MetaArgs)
	if ret != 0 {
		return ret
	}

	_ = packerStarter.Initialize(packer.InitializeOptions{})

	// Determine if stdin is a pipe. If so, we evaluate directly.
	if c.StdinPiped() {
		return c.modePiped(packerStarter)
	}

	return c.modeInteractive(packerStarter)
}

func (*ConsoleCommand) Help() string {
	helpText := `
Usage: packer console [options] [TEMPLATE]

  Creates a console for testing variable interpolation.
  If a template is provided, this command will load the template and any
  variables defined therein into its context to be referenced during
  interpolation.

Options:
  -var 'key=value'       Variable for templates, can be used multiple times.
  -var-file=path         JSON or HCL2 file containing user variables.
  -config-type           Set to 'hcl2' to run in HCL2 mode when no file is passed. Defaults to json.
`

	return strings.TrimSpace(helpText)
}

func (*ConsoleCommand) Synopsis() string {
	return "creates a console for testing variable interpolation"
}

func (*ConsoleCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*ConsoleCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-var":      complete.PredictNothing,
		"-var-file": complete.PredictNothing,
	}
}

func (c *ConsoleCommand) modePiped(cfg packer.Evaluator) int {
	var lastResult string
	scanner := bufio.NewScanner(wrappedstreams.Stdin())
	ret := 0
	for scanner.Scan() {
		result, _, diags := cfg.EvaluateExpression(strings.TrimSpace(scanner.Text()))
		if len(diags) > 0 {
			ret = writeDiags(c.Ui, nil, diags)
		}
		// Store the last result
		lastResult = result
	}

	// Output the final result
	c.Ui.Message(lastResult)
	return ret
}

func (c *ConsoleCommand) modeInteractive(cfg packer.Evaluator) int {
	// Setup the UI so we can output directly to stdout
	l, err := readline.NewEx(wrappedreadline.Override(&readline.Config{
		Prompt:            "> ",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	}))
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error initializing console: %s",
			err))
		return 1
	}
	for {
		// Read a line
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		out, exit, diags := cfg.EvaluateExpression(line)
		ret := writeDiags(c.Ui, nil, diags)
		if exit {
			return ret
		}
		c.Ui.Say(out)
		if exit {
			return ret
		}
	}

	return 0
}
