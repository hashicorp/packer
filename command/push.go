package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/packer/packer"
)

type PushCommand struct {
	Meta
}

func (c *PushCommand) Run(args []string) int {
	f := flag.NewFlagSet("push", flag.ContinueOnError)
	f.Usage = func() { c.Ui.Error(c.Help()) }
	if err := f.Parse(args); err != nil {
		return 1
	}

	args = f.Args()
	if len(args) != 1 {
		f.Usage()
		return 1
	}

	// Read the template
	tpl, err := packer.ParseTemplateFile(args[0], nil)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	// TODO: validate the template
	println(tpl.Push.Name)

	return 0
}

func (*PushCommand) Help() string {
	helpText := `
Usage: packer push [options] TEMPLATE

  Push the template and the files it needs to a Packer build service.
  This will not initiate any builds, it will only update the templates
  used for builds.
`

	return strings.TrimSpace(helpText)
}

func (*PushCommand) Synopsis() string {
	return "push template files to a Packer build service"
}
