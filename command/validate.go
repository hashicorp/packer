package command

import (
	"flag"
	"fmt"
	cmdcommon "github.com/mitchellh/packer/common/command"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
)

type ValidateCommand struct {
	Meta
}

func (c *ValidateCommand) Run(args []string) int {
	var cfgSyntaxOnly bool
	buildOptions := new(cmdcommon.BuildOptions)

	env, err := c.Meta.Environment()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing environment: %s", err))
		return 1
	}

	cmdFlags := flag.NewFlagSet("validate", flag.ContinueOnError)
	cmdFlags.Usage = func() { env.Ui().Say(c.Help()) }
	cmdFlags.BoolVar(&cfgSyntaxOnly, "syntax-only", false, "check syntax only")
	cmdcommon.BuildOptionFlags(cmdFlags, buildOptions)
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	args = cmdFlags.Args()
	if len(args) != 1 {
		cmdFlags.Usage()
		return 1
	}

	if err := buildOptions.Validate(); err != nil {
		env.Ui().Error(err.Error())
		env.Ui().Error("")
		env.Ui().Error(c.Help())
		return 1
	}

	userVars, err := buildOptions.AllUserVars()
	if err != nil {
		env.Ui().Error(fmt.Sprintf("Error compiling user variables: %s", err))
		env.Ui().Error("")
		env.Ui().Error(c.Help())
		return 1
	}

	// Parse the template into a machine-usable format
	log.Printf("Reading template: %s", args[0])
	tpl, err := packer.ParseTemplateFile(args[0], userVars)
	if err != nil {
		env.Ui().Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	if cfgSyntaxOnly {
		env.Ui().Say("Syntax-only check passed. Everything looks okay.")
		return 0
	}

	errs := make([]error, 0)
	warnings := make(map[string][]string)

	// The component finder for our builds
	components := &packer.ComponentFinder{
		Builder:       env.Builder,
		Hook:          env.Hook,
		PostProcessor: env.PostProcessor,
		Provisioner:   env.Provisioner,
	}

	// Otherwise, get all the builds
	builds, err := buildOptions.Builds(tpl, components)
	if err != nil {
		env.Ui().Error(err.Error())
		return 1
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
		env.Ui().Error("Template validation failed. Errors are shown below.\n")
		for i, err := range errs {
			env.Ui().Error(err.Error())

			if (i + 1) < len(errs) {
				env.Ui().Error("")
			}
		}

		return 1
	}

	if len(warnings) > 0 {
		env.Ui().Say("Template validation succeeded, but there were some warnings.")
		env.Ui().Say("These are ONLY WARNINGS, and Packer will attempt to build the")
		env.Ui().Say("template despite them, but they should be paid attention to.\n")

		for build, warns := range warnings {
			env.Ui().Say(fmt.Sprintf("Warnings for build '%s':\n", build))
			for _, warning := range warns {
				env.Ui().Say(fmt.Sprintf("* %s", warning))
			}
		}

		return 0
	}

	env.Ui().Say("Template validated successfully.")
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
