package command

import (
	"context"
	"os"
	"strings"

	hclutils "github.com/hashicorp/packer/hcl2template"
	"github.com/posener/complete"
)

type FormatCommand struct {
	Meta
}

func (c *FormatCommand) Run(args []string) int {
	ctx := context.Background()
	cfg, ret := c.ParseArgs(args)
	if ret != 0 {
		return ret
	}

	return c.RunContext(ctx, cfg)
}

func (c *FormatCommand) ParseArgs(args []string) (*FormatArgs, int) {
	var cfg FormatArgs
	flags := c.Meta.FlagSet("format", FlagSetNone)
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

func (c *FormatCommand) RunContext(ctx context.Context, cla *FormatArgs) int {
	if cla.Check {
		cla.Write = false
	}

	formatter := hclutils.HCL2Formatter{
		ShowDiff: cla.Diff,
		Write:    cla.Write,
		Output:   os.Stdout,
	}

	bytesModified, diags := formatter.Format(cla.Path)
	ret := writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	if cla.Check && bytesModified > 0 {
		return 3
	}

	return 0
}

func (*FormatCommand) Help() string {
	helpText := `
Usage: packer fmt [options] [TEMPLATE]

  Rewrites all Packer configuration files to a canonical format. Both
  configuration files (.pkr.hcl) and variable files (.pkrvars) are updated.
  JSON files (.json) are not modified.

  If TEMPATE is "." the current directory will be used. The given content must
  be in Packer's HCL2 configuration language; JSON is not supported.

Options:
  -check        Check if the input is formatted. Exit status will be 0 if all
                 input is properly formatted and non-zero otherwise.

  -diff         Display diffs of formatting change

  -write=false  Don't write to source files
                (always disabled if using -check)

`

	return strings.TrimSpace(helpText)
}

func (*FormatCommand) Synopsis() string {
	return "Rewrites HCL2 config files to canonical format"
}

func (*FormatCommand) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (*FormatCommand) AutocompleteFlags() complete.Flags {
	return complete.Flags{
		"-check": complete.PredictNothing,
		"-diff":  complete.PredictNothing,
		"-write": complete.PredictNothing,
	}
}
