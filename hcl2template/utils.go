package hcl2template

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template/repl"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/zclconf/go-cty/cty"
)

func warningErrorsToDiags(block *hcl.Block, warnings []string, err error) hcl.Diagnostics {
	var diags hcl.Diagnostics

	for _, warning := range warnings {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  warning,
			Subject:  &block.DefRange,
			Severity: hcl.DiagWarning,
		})
	}
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  err.Error(),
			Subject:  &block.DefRange,
			Severity: hcl.DiagError,
		})
	}
	return diags
}

func isDir(name string) (bool, error) {
	s, err := os.Stat(name)
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

// GetHCL2Files returns two slices of json formatted and hcl formatted files,
// hclSuffix and jsonSuffix tell which file is what. Filename can be a folder
// or a file.
//
// When filename is a folder all files of folder matching the suffixes will be
// returned. Otherwise if filename references a file and filename matches one
// of the suffixes it is returned in the according slice.
func GetHCL2Files(filename, hclSuffix, jsonSuffix string) (hclFiles, jsonFiles []string, diags hcl.Diagnostics) {
	if filename == "" {
		return
	}
	isDir, err := isDir(filename)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Cannot tell wether " + filename + " is a directory",
			Detail:   err.Error(),
		})
		return nil, nil, diags
	}
	if !isDir {
		if strings.HasSuffix(filename, jsonSuffix) {
			return nil, []string{filename}, diags
		}
		if strings.HasSuffix(filename, hclSuffix) {
			return []string{filename}, nil, diags
		}
		return nil, nil, diags
	}

	fileInfos, err := ioutil.ReadDir(filename)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Cannot read hcl directory",
			Detail:   err.Error(),
		}
		diags = append(diags, diag)
		return nil, nil, diags
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}
		filename := filepath.Join(filename, fileInfo.Name())
		if strings.HasSuffix(filename, hclSuffix) {
			hclFiles = append(hclFiles, filename)
		} else if strings.HasSuffix(filename, jsonSuffix) {
			jsonFiles = append(jsonFiles, filename)
		}
	}

	return hclFiles, jsonFiles, diags
}

// Convert -only and -except globs to glob.Glob instances.
func convertFilterOption(patterns []string, optionName string) ([]glob.Glob, hcl.Diagnostics) {
	var globs []glob.Glob
	var diags hcl.Diagnostics

	for _, pattern := range patterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Summary:  fmt.Sprintf("Invalid -%s pattern %s: %s", optionName, pattern, err),
				Severity: hcl.DiagError,
			})
		}
		globs = append(globs, g)
	}

	return globs, diags
}

func PrintableCtyValue(v cty.Value) string {
	if !v.IsWhollyKnown() {
		return "<unknown>"
	}
	gval := hcl2shim.ConfigValueFromHCL2(v)
	str := repl.FormatResult(gval)
	return str
}

func ConvertPluginConfigValueToHCLValue(v interface{}) (cty.Value, error) {
	var buildValue cty.Value
	switch v := v.(type) {
	case bool:
		buildValue = cty.BoolVal(v)
	case string:
		buildValue = cty.StringVal(v)
	case uint8:
		buildValue = cty.NumberUIntVal(uint64(v))
	case float64:
		buildValue = cty.NumberFloatVal(v)
	case int64:
		buildValue = cty.NumberIntVal(v)
	case uint64:
		buildValue = cty.NumberUIntVal(v)
	case []string:
		vals := make([]cty.Value, len(v))
		for i, ev := range v {
			vals[i] = cty.StringVal(ev)
		}
		if len(vals) == 0 {
			buildValue = cty.ListValEmpty(cty.String)
		} else {
			buildValue = cty.ListVal(vals)
		}
	case []uint8:
		vals := make([]cty.Value, len(v))
		for i, ev := range v {
			vals[i] = cty.NumberUIntVal(uint64(ev))
		}
		if len(vals) == 0 {
			buildValue = cty.ListValEmpty(cty.Number)
		} else {
			buildValue = cty.ListVal(vals)
		}
	case []int64:
		vals := make([]cty.Value, len(v))
		for i, ev := range v {
			vals[i] = cty.NumberIntVal(ev)
		}
		if len(vals) == 0 {
			buildValue = cty.ListValEmpty(cty.Number)
		} else {
			buildValue = cty.ListVal(vals)
		}
	case []uint64:
		vals := make([]cty.Value, len(v))
		for i, ev := range v {
			vals[i] = cty.NumberUIntVal(ev)
		}
		if len(vals) == 0 {
			buildValue = cty.ListValEmpty(cty.Number)
		} else {
			buildValue = cty.ListVal(vals)
		}
	default:
		return cty.Value{}, fmt.Errorf("unhandled buildvar type: %T", v)
	}
	return buildValue, nil
}
