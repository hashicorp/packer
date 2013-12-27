package inspect

import (
	"flag"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"log"
	"sort"
	"strings"
)

type Command struct{}

func (Command) Help() string {
	return strings.TrimSpace(helpText)
}

func (c Command) Synopsis() string {
	return "see components of a template"
}

func (c Command) Run(env packer.Environment, args []string) int {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	flags.Usage = func() { env.Ui().Say(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return 1
	}

	// Read the file into a byte array so that we can parse the template
	log.Printf("Reading template: %#v", args[0])
	tpl, err := packer.ParseTemplateFile(args[0], nil)
	if err != nil {
		env.Ui().Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	// Convenience...
	ui := env.Ui()

	// Description
	if tpl.Description != "" {
		ui.Say("Description:\n")
		ui.Say(tpl.Description + "\n")
	}

	// Variables
	if len(tpl.Variables) == 0 {
		ui.Say("Variables:\n")
		ui.Say("  <No variables>")
	} else {
		requiredHeader := false
		for k, v := range tpl.Variables {
			if v.Required {
				if !requiredHeader {
					requiredHeader = true
					ui.Say("Required variables:\n")
				}

				ui.Machine("template-variable", k, v.Default, "1")
				ui.Say("  " + k)
			}
		}

		if requiredHeader {
			ui.Say("")
		}

		ui.Say("Optional variables and their defaults:\n")
		keys := make([]string, 0, len(tpl.Variables))
		max := 0
		for k, _ := range tpl.Variables {
			keys = append(keys, k)
			if len(k) > max {
				max = len(k)
			}
		}

		sort.Strings(keys)

		for _, k := range keys {
			v := tpl.Variables[k]
			if v.Required {
				continue
			}

			padding := strings.Repeat(" ", max-len(k))
			output := fmt.Sprintf("  %s%s = %s", k, padding, v.Default)

			ui.Machine("template-variable", k, v.Default, "0")
			ui.Say(output)
		}
	}

	ui.Say("")

	// Builders
	ui.Say("Builders:\n")
	if len(tpl.Builders) == 0 {
		ui.Say("  <No builders>")
	} else {
		keys := make([]string, 0, len(tpl.Builders))
		max := 0
		for k, _ := range tpl.Builders {
			keys = append(keys, k)
			if len(k) > max {
				max = len(k)
			}
		}

		sort.Strings(keys)

		for _, k := range keys {
			v := tpl.Builders[k]
			padding := strings.Repeat(" ", max-len(k))
			output := fmt.Sprintf("  %s%s", k, padding)
			if v.Name != v.Type {
				output = fmt.Sprintf("%s (%s)", output, v.Type)
			}

			ui.Machine("template-builder", k, v.Type)
			ui.Say(output)

		}
	}

	ui.Say("")

	// Provisioners
	ui.Say("Provisioners:\n")
	if len(tpl.Provisioners) == 0 {
		ui.Say("  <No provisioners>")
	} else {
		for _, v := range tpl.Provisioners {
			ui.Machine("template-provisioner", v.Type)
			ui.Say(fmt.Sprintf("  %s", v.Type))
		}
	}

	ui.Say("\nNote: If your build names contain user variables or template\n" +
		"functions such as 'timestamp', these are processed at build time,\n" +
		"and therefore only show in their raw form here.")

	return 0
}
