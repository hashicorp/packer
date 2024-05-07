// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer/fix"

	"github.com/posener/complete"
)

type FixCommand struct {
	Meta
}

func (c *FixCommand) Run(args []string) int {
	ctx, cleanup := handleTermInterrupt(c.Ui)
	defer cleanup()

	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *FixCommand) ParseArgs(args []string) (*FixArgs, int) {
	var cfg FixArgs
	flags := c.Meta.FlagSet("fix")
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

func (c *FixCommand) RunContext(ctx context.Context, cla *FixArgs) int {
	if hcl2, _ := isHCLLoaded(cla.Path); hcl2 {
		c.Ui.Error("packer fix only works with JSON files for now.")
		return 1
	}
	// Read the file for decoding
	tplF, err := os.Open(cla.Path)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error opening template: %s", err))
		return 1
	}
	defer tplF.Close()

	// Decode the JSON into a generic map structure
	var templateData map[string]interface{}
	decoder := json.NewDecoder(tplF)
	if err := decoder.Decode(&templateData); err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing template: %s", err))
		return 1
	}

	// Close the file since we're done with that
	tplF.Close()

	input := templateData
	for _, name := range fix.FixerOrder {
		var err error
		fixer, ok := fix.Fixers[name]
		if !ok {
			panic("fixer not found: " + name)
		}

		log.Printf("Running fixer: %s", name)
		input, err = fixer.Fix(input)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error fixing: %s", err))
			return 1
		}
	}

	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	if err := encoder.Encode(input); err != nil {
		c.Ui.Error(fmt.Sprintf("Error encoding: %s", err))
		return 1
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, output.Bytes(), "", "  "); err != nil {
		c.Ui.Error(fmt.Sprintf("Error encoding: %s", err))
		return 1
	}

	result := indented.String()
	result = strings.Replace(result, `\u003c`, "<", -1)
	result = strings.Replace(result, `\u003e`, ">", -1)
	c.Ui.Say(result)

	if cla.Validate == false {
		return 0
	}

	// Attempt to parse and validate the template
	tpl, err := template.Parse(strings.NewReader(result))
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error! Fixed template fails to parse: %s\n\n"+
				"This is usually caused by an error in the input template.\n"+
				"Please fix the error and try again.",
			err))
		return 1
	}
	if err := tpl.Validate(); err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error! Fixed template failed to validate: %s\n\n"+
				"This is usually caused by an error in the input template.\n"+
				"Please fix the error and try again.",
			err))
		return 1
	}

	return 0
}

func (*FixCommand) Help() string {
	helpText := `
Usage: packer fix [options] TEMPLATE

  Reads the JSON template and attempts to fix known backwards
  incompatibilities. The fixed template will be outputted to standard out.

  If the template cannot be fixed due to an error, the command will exit
  with a non-zero exit status. Error messages will appear on standard error.

Fixes that are run (in order):

`

	for _, name := range fix.FixerOrder {
		helpText += fmt.Sprintf(
			"  %-27s%s\n", name, fix.Fixers[name].Synopsis())
	}

	helpText += `
Options:

  -validate=true      If true (default), validates the fixed template.
`

	return strings.TrimSpace(helpText)
}

func (c *FixCommand) Synopsis() string {
	return "fixes templates from old versions of packer"
}

func (c *FixCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (c *FixCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-validate": complete.PredictNothing,
	}
}
