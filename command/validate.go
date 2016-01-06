package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template"
)

type ValidateCommand struct {
	Meta
}

func (c *ValidateCommand) Run(args []string) int {
	var cfgSyntaxOnly bool
	flags := c.Meta.FlagSet("validate", FlagSetBuildFilter|FlagSetVars)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	flags.BoolVar(&cfgSyntaxOnly, "syntax-only", false, "check syntax only")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return 1
	}

	// Parse the template
	tpl, err := template.ParseFile(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	// If we're only checking syntax, then we're done already
	if cfgSyntaxOnly {
		c.Ui.Say("Syntax-only check passed. Everything looks okay.")
		return 0
	}

	// Get the core
	core, err := c.Meta.Core(tpl)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	errs := make([]error, 0)
	warnings := make(map[string][]string)

	// Get the builds we care about
	buildNames := c.Meta.BuildNames(core)
	builds := make([]packer.Build, 0, len(buildNames))
	for _, n := range buildNames {
		b, err := core.Build(n)
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Failed to initialize build '%s': %s",
				n, err))
			return 1
		}

		builds = append(builds, b)
	}

	// Check the configuration of all builds
	for _, b := range builds {
		log.Printf("Preparing build: %s", b.Name())
		warns, err := b.Prepare()
		if len(warns) > 0 {
			warnings[b.Name()] = warns
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("Errors validating build '%s'. %s", b.Name(), err))
		}
	}

	if len(errs) > 0 {
		c.Ui.Error("Template validation failed. Errors are shown below.\n")
		for i, err := range errs {
			c.Ui.Error(err.Error())

			if (i + 1) < len(errs) {
				c.Ui.Error("")
			}
		}

		return 1
	}

	if len(warnings) > 0 {
		c.Ui.Say("Template validation succeeded, but there were some warnings.")
		c.Ui.Say("These are ONLY WARNINGS, and Packer will attempt to build the")
		c.Ui.Say("template despite them, but they should be paid attention to.\n")

		for build, warns := range warnings {
			c.Ui.Say(fmt.Sprintf("Warnings for build '%s':\n", build))
			for _, warning := range warns {
				c.Ui.Say(fmt.Sprintf("* %s", warning))
			}
		}

		return 0
	}

	c.Ui.Say("Template validated successfully.")
	return 0
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

  -syntax-only           Only check syntax. Do not verify config of the template.
  -except=foo,bar,baz    Validate all builds other than these
  -only=foo,bar,baz      Validate only these builds
  -var 'key=value'       Variable for templates, can be used multiple times.
  -var-file=path         JSON file containing user variables.
`

	return strings.TrimSpace(helpText)
}

func (*ValidateCommand) Synopsis() string {
	return "check that a template is valid"
}
