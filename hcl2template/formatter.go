package hcl2template

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type HCL2Formatter struct {
	ShowDiff, Write bool
	Output          io.Writer
	parser          *hclparse.Parser
}

// NewHCL2Formatter creates a new formatter, ready to format configuration files.
func NewHCL2Formatter() *HCL2Formatter {
	return &HCL2Formatter{
		parser: hclparse.NewParser(),
	}
}

// Format all HCL2 files in path and return the total bytes formatted.
// If any error is encountered, zero bytes will be returned.
//
// Path can be a directory or a file.
func (f *HCL2Formatter) Format(path string) (int, hcl.Diagnostics) {

	var allHclFiles []string
	var diags []*hcl.Diagnostic

	if path == "-" {
		allHclFiles = []string{"-"}
	} else {
		hclFiles, _, diags := GetHCL2Files(path, hcl2FileExt, hcl2JsonFileExt)
		if diags.HasErrors() {
			return 0, diags
		}

		hclVarFiles, _, diags := GetHCL2Files(path, hcl2VarFileExt, hcl2VarJsonFileExt)
		if diags.HasErrors() {
			return 0, diags
		}

		allHclFiles = append(hclFiles, hclVarFiles...)

		if len(allHclFiles) == 0 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Cannot tell whether %s contains HCL2 configuration data", path),
			})

			return 0, diags
		}
	}

	if f.parser == nil {
		f.parser = hclparse.NewParser()
	}

	var bytesModified int
	for _, fn := range allHclFiles {
		data, err := f.processFile(fn)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("encountered an error while formatting %s", fn),
				Detail:   err.Error(),
			})
		}
		bytesModified += len(data)
	}

	return bytesModified, diags
}

// processFile formats the source contents of filename and return the formatted data.
// overwriting the contents of the original when the f.Write is true; a diff of the changes
// will be outputted if f.ShowDiff is true.
func (f *HCL2Formatter) processFile(filename string) ([]byte, error) {
	if f.Output == nil {
		f.Output = os.Stdout
	}

	var in io.Reader
	var err error

	if filename == "-" {
		in = os.Stdin
		f.ShowDiff = false
	} else {
		in, err = os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %s", filename, err)
		}
	}

	inSrc, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", filename, err)
	}

	_, diags := f.parser.ParseHCL(inSrc, filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL %s", filename)
	}

	outSrc := hclwrite.Format(inSrc)

	if bytes.Equal(inSrc, outSrc) {
		return nil, nil
	}

	if filename != "-" {
		s := []byte(fmt.Sprintf("%s\n", filename))
		_, _ = f.Output.Write(s)
	}

	if f.Write {
		if filename == "-" {
			_, _ = f.Output.Write(outSrc)
		} else {
			if err := ioutil.WriteFile(filename, outSrc, 0644); err != nil {
				return nil, err
			}
		}
	}

	if f.ShowDiff {
		diff, err := bytesDiff(inSrc, outSrc, filename)
		if err != nil {
			return outSrc, fmt.Errorf("failed to generate diff for %s: %s", filename, err)
		}
		_, _ = f.Output.Write(diff)
	}

	return outSrc, nil
}

// bytesDiff returns the unified diff of b1 and b2
// Shamelessly copied from Terraform's fmt command.
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
