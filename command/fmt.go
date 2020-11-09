package command

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
		c.Ui.Say("No HCL files found; please check that all HCL files end with the proper suffix.")
		return 0
	}

	if cla.Check {
		cla.Write = false
	}

	var bytesChanged int
	var err error
	c.parser = hclparse.NewParser()
	for _, path := range allHclFiles {
		bytesChanged, err = c.processFile(path, cla.Write, cla.Diff)
		if err != nil {
			c.Ui.Say(err.Error())
			return 1
		}

	}

	if cla.Check && bytesChanged != 0 {
		// exit code taken from Terraform fmt
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
		"-write": complete.PredictNothing,
	}
}

// processFile formats the source contents of filename if it is not properly formatted
// overwriting the contents of the original file if overwrite is set. A diff of the changes
// will be outputted if showDiff is true.
func (c *FormatCommand) processFile(filename string, overwrite bool, showDiff bool) (int, error) {

	in, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to open %s: %s", filename, err)
	}

	inSrc, err := ioutil.ReadAll(in)
	if err != nil {
		return 0, fmt.Errorf("failed to read %s: %s", filename, err)
	}

	_, diags := c.parser.ParseHCL(inSrc, filename)
	ret := writeDiags(c.Ui, nil, diags)
	if ret != 0 {
		return 0, fmt.Errorf("failed to parse HCL %s", filename)
	}

	outSrc := hclwrite.Format(inSrc)

	if bytes.Equal(inSrc, outSrc) {
		return 0, nil
	}

	// Display filename as we have changes
	c.Ui.Say(filename)
	if overwrite {
		if err := ioutil.WriteFile(filename, outSrc, 0644); err != nil {
			c.Ui.Say(err.Error())
			return 0, err
		}
	}

	if showDiff {
		diff, err := bytesDiff(inSrc, outSrc, filename)
		if err != nil {
			c.Ui.Say(fmt.Sprintf("failed to generate diff for %s: %s", filename, err))
			return len(outSrc), nil
		}
		_, _ = os.Stdout.Write(diff)
	}

	return len(outSrc), nil
}

func bytesDiff(b1, b2 []byte, path string) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "")
	if err != nil {
		return
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	_, _ = f1.Write(b1)
	_, _ = f2.Write(b2)

	data, err = exec.Command("diff", "--label=old/"+path, "--label=new/"+path, "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return
}
