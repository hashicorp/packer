package command

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/fix"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template"

	"github.com/google/go-cmp/cmp"
	"github.com/posener/complete"
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

	// Check if any of the configuration is fixable
	var rawTemplateData map[string]interface{}
	input := make(map[string]interface{})
	templateData := make(map[string]interface{})
	json.Unmarshal(tpl.RawContents, &rawTemplateData)
	for k, v := range rawTemplateData {
		if vals, ok := v.([]interface{}); ok {
			if len(vals) == 0 {
				continue
			}
		}
		templateData[strings.ToLower(k)] = v
		input[strings.ToLower(k)] = v
	}

	// fix rawTemplateData into input
	for _, name := range fix.FixerOrder {
		var err error
		fixer, ok := fix.Fixers[name]
		if !ok {
			panic("fixer not found: " + name)
		}
		input, err = fixer.Fix(input)
		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error checking against fixers: %s", err))
			return 1
		}
	}
	// delete empty top-level keys since the fixers seem to add them
	// willy-nilly
	for k := range input {
		ml, ok := input[k].([]map[string]interface{})
		if !ok {
			continue
		}
		if len(ml) == 0 {
			delete(input, k)
		}
	}
	// marshal/unmarshal to make comparable to templateData
	var fixedData map[string]interface{}
	// Guaranteed to be valid json, so we can ignore errors
	j, _ := json.Marshal(input)
	json.Unmarshal(j, &fixedData)

	if diff := cmp.Diff(templateData, fixedData); diff != "" {
		c.Ui.Say("[warning] Fixable configuration found.")
		c.Ui.Say("You may need to run `packer fix` to get your build to run")
		c.Ui.Say("correctly. See debug log for more information.\n")
		log.Printf("Fixable config differences:\n%s", diff)
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

func (*ValidateCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*ValidateCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-syntax-only": complete.PredictNothing,
		"-except":      complete.PredictNothing,
		"-only":        complete.PredictNothing,
		"-var":         complete.PredictNothing,
		"-var-file":    complete.PredictNothing,
	}
}
