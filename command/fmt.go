package command

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	hclutils "github.com/hashicorp/packer/hcl2template"
	"github.com/posener/complete"
)

const (
	hcl2FileExt        = ".pkr.hcl"
	hcl2JsonFileExt    = ".pkr.json"
	hcl2VarFileExt     = ".auto.pkrvars.hcl"
	hcl2VarJsonFileExt = ".auto.pkrvars.json"
)

type FormatCommand struct {
	Meta
	parser *hclparse.Parser
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
	hclFiles, _, diags := hclutils.GetHCL2Files(cla.Path, hcl2FileExt, hcl2JsonFileExt)
	ret := writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	hclVarFiles, _, diags := hclutils.GetHCL2Files(cla.Path, hcl2VarFileExt, hcl2VarJsonFileExt)
	ret = writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return ret
	}

	allHclFiles := append(hclFiles, hclVarFiles...)

	if len(allHclFiles) == 0 {
		c.Ui.Say("No HCL files found; please check that all HCL files end with the proper suffix")
		return 0
	}

	c.parser = hclparse.NewParser()
	for _, path := range allHclFiles {
		b, err := c.formatFile(path)
		if err != nil {
			c.Ui.Say(err.Error())
			return 1
		}

		if b == nil {
			continue
		}

		c.Ui.Say(path)
		if cla.Check {
			continue
		}

		if err := ioutil.WriteFile(path, b, 0644); err != nil {
			c.Ui.Say(err.Error())
			return 1
		}

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
  -check=false  Check if the input is formatted. Exit status will be 0 if all
                 input is properly formatted and non-zero otherwise.

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
		"-write": complete.PredictNothing,
	}
}

// formatFile formats the source context of filename if it is not properly formatted.
func (c *FormatCommand) formatFile(filename string) ([]byte, error) {

	in, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %s", filename, err)
	}

	inSrc, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", filename, err)
	}

	_, diags := c.parser.ParseHCL(inSrc, filename)
	ret := writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return nil, fmt.Errorf("failed to parse HCL %s", filename)
	}

	outSrc := hclwrite.Format(inSrc)

	if bytes.Equal(inSrc, outSrc) {
		return nil, nil
	}

	//_, err = os.Stdout.Write(outSrc)
	//
	return outSrc, nil
}
