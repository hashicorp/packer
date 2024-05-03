// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type HCL2Formatter struct {
	ShowDiff, Write, Recursive bool
	Output                     io.Writer
	parser                     *hclparse.Parser
}

// NewHCL2Formatter creates a new formatter, ready to format configuration files.
func NewHCL2Formatter() *HCL2Formatter {
	return &HCL2Formatter{
		parser: hclparse.NewParser(),
	}
}

func isHcl2FileOrVarFile(path string) bool {
	if strings.HasSuffix(path, hcl2FileExt) || strings.HasSuffix(path, hcl2VarFileExt) {
		return true
	}
	return false
}

func (f *HCL2Formatter) formatFile(path string, diags hcl.Diagnostics, bytesModified int) (int, hcl.Diagnostics) {
	data, err := f.processFile(path)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("encountered an error while formatting %s", path),
			Detail:   err.Error(),
		})
	}
	bytesModified += len(data)
	return bytesModified, diags
}

// Format all HCL2 files in path and return the total bytes formatted.
// If any error is encountered, zero bytes will be returned.
//
// Path can be a directory or a file.
func (f *HCL2Formatter) Format(path string) (int, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var bytesModified int

	if path == "" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "path is empty, cannot format",
			Detail:   "path is empty, cannot format",
		})
		return bytesModified, diags
	}

	if f.parser == nil {
		f.parser = hclparse.NewParser()
	}

	if s, err := os.Stat(path); err != nil || !s.IsDir() {
		return f.formatFile(path, diags, bytesModified)
	}

	fileInfos, err := os.ReadDir(path)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Cannot read hcl directory",
			Detail:   err.Error(),
		}
		diags = append(diags, diag)
		return bytesModified, diags
	}

	for _, fileInfo := range fileInfos {
		filename := filepath.Join(path, fileInfo.Name())
		if fileInfo.IsDir() {
			if f.Recursive {
				var tempDiags hcl.Diagnostics
				var tempBytesModified int
				tempBytesModified, tempDiags = f.Format(filename)
				bytesModified += tempBytesModified
				diags = diags.Extend(tempDiags)
			}
			continue
		}
		if isHcl2FileOrVarFile(filename) {
			bytesModified, diags = f.formatFile(filename, diags, bytesModified)
		}
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

	inSrc, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", filename, err)
	}

	_, diags := f.parser.ParseHCL(inSrc, filename)
	if diags.HasErrors() {
		return nil, multierror.Append(nil, diags.Errs()...)
	}

	outSrc := hclwrite.Format(inSrc)

	if bytes.Equal(inSrc, outSrc) {
		if filename == "-" {
			_, _ = f.Output.Write(outSrc)
		}

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
			if err := os.WriteFile(filename, outSrc, 0644); err != nil {
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
	f1, err := os.CreateTemp("", "")
	if err != nil {
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := os.CreateTemp("", "")
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
