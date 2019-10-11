package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/packer/fix"
	"github.com/hashicorp/packer/template"

	"github.com/posener/complete"
)

type FixCommand struct {
	Meta
}

func (c *FixCommand) Run(args []string) int {
	var flagValidate bool
	flags := c.Meta.FlagSet("fix", FlagSetNone)
	flags.BoolVar(&flagValidate, "validate", true, "")
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return 1
	}

	// Read the file for decoding
	tplF, err := os.Open(args[0])
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

	if flagValidate {
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
