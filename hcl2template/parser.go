// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
	"github.com/hashicorp/hcl/v2/hclparse"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

const (
	packerLabel       = "packer"
	sourceLabel       = "source"
	variablesLabel    = "variables"
	variableLabel     = "variable"
	localsLabel       = "locals"
	localLabel        = "local"
	dataSourceLabel   = "data"
	buildLabel        = "build"
	communicatorLabel = "communicator"
)

var configSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: packerLabel},
		{Type: sourceLabel, LabelNames: []string{"type", "name"}},
		{Type: variablesLabel},
		{Type: variableLabel, LabelNames: []string{"name"}},
		{Type: localsLabel},
		{Type: localLabel, LabelNames: []string{"name"}},
		{Type: dataSourceLabel, LabelNames: []string{"type", "name"}},
		{Type: buildLabel},
		{Type: communicatorLabel, LabelNames: []string{"type", "name"}},
	},
}

// packerBlockSchema is the schema for a top-level "packer" block in
// a configuration file.
var packerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "required_version"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "required_plugins"},
	},
}

// Parser helps you parse HCL folders. It will parse an hcl file or directory
// and start builders, provisioners and post-processors to configure them with
// the parsed HCL and then return a []packersdk.Build. Packer will use that list
// of Builds to run everything in order.
type Parser struct {
	CorePackerVersion *version.Version

	CorePackerVersionString string

	PluginConfig *packer.PluginConfig

	ValidationOptions

	*hclparse.Parser
}

const (
	hcl2FileExt            = ".pkr.hcl"
	hcl2JsonFileExt        = ".pkr.json"
	hcl2VarFileExt         = ".pkrvars.hcl"
	hcl2VarJsonFileExt     = ".pkrvars.json"
	hcl2AutoVarFileExt     = ".auto.pkrvars.hcl"
	hcl2AutoVarJsonFileExt = ".auto.pkrvars.json"
)

// Parse will Parse all HCL files in filename. Path can be a folder or a file.
//
// Parse will first Parse packer and variables blocks, omitting the rest, which
// can be expanded with dynamic blocks. We need to evaluate all variables for
// that, so that data sources can expand dynamic blocks too.
//
// Parse returns a PackerConfig that contains configuration layout of a packer
// build; sources(builders)/provisioners/posts-processors will not be started
// and their contents wont be verified; Most syntax errors will cause an error,
// init should be called next to expand dynamic blocks and verify that used
// things do exist.
func (p *Parser) Parse(filename string, varFiles []string, argVars map[string]string) (*PackerConfig, hcl.Diagnostics) {
	var files []*hcl.File
	var diags hcl.Diagnostics

	// parse config files
	if filename != "" {
		hclFiles, jsonFiles, moreDiags := GetHCL2Files(filename, hcl2FileExt, hcl2JsonFileExt)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			// here this probably means that the file was not found, let's
			// simply leave early.
			return nil, diags
		}
		if len(hclFiles)+len(jsonFiles) == 0 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Could not find any config file in " + filename,
				Detail: "A config file must be suffixed with `.pkr.hcl` or " +
					"`.pkr.json`. A folder can be referenced.",
			})
		}
		for _, filename := range hclFiles {
			f, moreDiags := p.ParseHCLFile(filename)
			diags = append(diags, moreDiags...)
			files = append(files, f)
		}
		for _, filename := range jsonFiles {
			f, moreDiags := p.ParseJSONFile(filename)
			diags = append(diags, moreDiags...)
			files = append(files, f)
		}
		if diags.HasErrors() {
			return nil, diags
		}
	}

	basedir := filename
	if isDir, err := isDir(basedir); err == nil && !isDir {
		basedir = filepath.Dir(basedir)
	}
	wd, err := os.Getwd()
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Could not find current working directory",
			Detail:   err.Error(),
		})
	}
	cfg := &PackerConfig{
		Basedir:                 basedir,
		Cwd:                     wd,
		CorePackerVersionString: p.CorePackerVersionString,
		HCPVars:                 map[string]cty.Value{},
		ValidationOptions:       p.ValidationOptions,
		parser:                  p,
		files:                   files,
	}

	for _, file := range files {
		coreVersionConstraints, moreDiags := sniffCoreVersionRequirements(file.Body)
		cfg.Packer.VersionConstraints = append(cfg.Packer.VersionConstraints, coreVersionConstraints...)
		diags = append(diags, moreDiags...)
	}

	// Before we go further, we'll check to make sure this version can read
	// all files, so we can produce a version-related error message rather than
	// potentially-confusing downstream errors.
	versionDiags := cfg.CheckCoreVersionRequirements(p.CorePackerVersion.Core())
	diags = append(diags, versionDiags...)
	if versionDiags.HasErrors() {
		return cfg, diags
	}

	// Decode required_plugins blocks.
	//
	// Note: using `latest` ( or actually an empty string ) in a config file
	// does not work and packer will ask you to pick a version
	{
		for _, file := range files {
			diags = append(diags, cfg.decodeRequiredPluginsBlock(file)...)
		}
	}

	// Decode variable blocks so that they are available later on. Here locals
	// can use input variables so we decode input variables first.
	{
		for _, file := range files {
			diags = append(diags, cfg.decodeInputVariables(file)...)
		}

		for _, file := range files {
			morediags := p.decodeDatasources(file, cfg)
			diags = append(diags, morediags...)
		}

		for _, file := range files {
			moreLocals, morediags := parseLocalVariableBlocks(file)
			diags = append(diags, morediags...)
			cfg.LocalBlocks = append(cfg.LocalBlocks, moreLocals...)
		}
	}

	// parse var files
	{
		hclVarFiles, jsonVarFiles, moreDiags := GetHCL2Files(filename, hcl2AutoVarFileExt, hcl2AutoVarJsonFileExt)
		diags = append(diags, moreDiags...)
		for _, file := range varFiles {
			switch filepath.Ext(file) {
			case ".hcl":
				hclVarFiles = append(hclVarFiles, file)
			case ".json":
				jsonVarFiles = append(jsonVarFiles, file)
			default:
				diags = append(moreDiags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Could not guess format of " + file,
					Detail:   "A var file must be suffixed with `.hcl` or `.json`.",
				})
			}
		}
		var varFiles []*hcl.File
		for _, filename := range hclVarFiles {
			f, moreDiags := p.ParseHCLFile(filename)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			varFiles = append(varFiles, f)
		}
		for _, filename := range jsonVarFiles {
			f, moreDiags := p.ParseJSONFile(filename)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			varFiles = append(varFiles, f)
		}

		diags = append(diags, cfg.collectInputVariableValues(os.Environ(), varFiles, argVars)...)
	}

	return cfg, diags
}

// sniffCoreVersionRequirements does minimal parsing of the given body for
// "packer" blocks with "required_version" attributes, returning the
// requirements found.
//
// This is intended to maximize the chance that we'll be able to read the
// requirements (syntax errors notwithstanding) even if the config file contains
// constructs that might've been added in future versions
//
// This is a "best effort" sort of method which will return constraints it is
// able to find, but may return no constraints at all if the given body is
// so invalid that it cannot be decoded at all.
func sniffCoreVersionRequirements(body hcl.Body) ([]VersionConstraint, hcl.Diagnostics) {

	var sniffRootSchema = &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: packerLabel,
			},
		},
	}

	rootContent, _, diags := body.PartialContent(sniffRootSchema)

	var constraints []VersionConstraint

	for _, block := range rootContent.Blocks {
		content, blockDiags := block.Body.Content(packerBlockSchema)
		diags = append(diags, blockDiags...)

		attr, exists := content.Attributes["required_version"]
		if !exists {
			continue
		}

		constraint, constraintDiags := decodeVersionConstraint(attr)
		diags = append(diags, constraintDiags...)
		if !constraintDiags.HasErrors() {
			constraints = append(constraints, constraint)
		}
	}

	return constraints, diags
}

func filterVarsFromLogs(inputOrLocal Variables) {
	for _, variable := range inputOrLocal {
		if !variable.Sensitive {
			continue
		}
		value := variable.Value()
		_ = cty.Walk(value, func(_ cty.Path, nested cty.Value) (bool, error) {
			if nested.IsWhollyKnown() && !nested.IsNull() && nested.Type().Equals(cty.String) {
				packersdk.LogSecretFilter.Set(nested.AsString())
			}
			return true, nil
		})
	}
}

func (cfg *PackerConfig) Initialize(opts packer.InitializeOptions) hcl.Diagnostics {
	diags := cfg.InputVariables.ValidateValues()
	diags = append(diags, cfg.evaluateDatasources(opts.SkipDatasourcesExecution)...)
	diags = append(diags, checkForDuplicateLocalDefinition(cfg.LocalBlocks)...)
	diags = append(diags, cfg.evaluateLocalVariables(cfg.LocalBlocks)...)

	filterVarsFromLogs(cfg.InputVariables)
	filterVarsFromLogs(cfg.LocalVariables)

	// parse the actual content // rest
	for _, file := range cfg.files {
		diags = append(diags, cfg.parser.parseConfig(file, cfg)...)
	}

	diags = append(diags, cfg.initializeBlocks()...)

	return diags
}

// parseConfig looks in the found blocks for everything that is not a variable
// block.
func (p *Parser) parseConfig(f *hcl.File, cfg *PackerConfig) hcl.Diagnostics {
	var diags hcl.Diagnostics

	body := f.Body
	body = dynblock.Expand(body, cfg.EvalContext(DatasourceContext, nil))
	content, moreDiags := body.Content(configSchema)
	diags = append(diags, moreDiags...)

	for _, block := range content.Blocks {
		switch block.Type {
		case sourceLabel:
			source, moreDiags := p.decodeSource(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			ref := source.Ref()
			if existing, found := cfg.Sources[ref]; found {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate " + sourceLabel + " block",
					Detail: fmt.Sprintf("This "+sourceLabel+" block has the "+
						"same builder type and name as a previous block declared "+
						"at %s. Each "+sourceLabel+" must have a unique name per builder type.",
						existing.block.DefRange.Ptr()),
					Subject: source.block.DefRange.Ptr(),
				})
				continue
			}

			if cfg.Sources == nil {
				cfg.Sources = map[SourceRef]SourceBlock{}
			}
			cfg.Sources[ref] = source

		case buildLabel:
			build, moreDiags := p.decodeBuildConfig(block, cfg)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			cfg.Builds = append(cfg.Builds, build)
		}
	}

	return diags
}

func (p *Parser) decodeDatasources(file *hcl.File, cfg *PackerConfig) hcl.Diagnostics {
	var diags hcl.Diagnostics

	body := file.Body
	content, moreDiags := body.Content(configSchema)
	diags = append(diags, moreDiags...)

	for _, block := range content.Blocks {
		switch block.Type {
		case dataSourceLabel:
			datasource, moreDiags := p.decodeDataBlock(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			ref := datasource.Ref()
			if existing, found := cfg.Datasources[ref]; found {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Duplicate " + dataSourceLabel + " block",
					Detail: fmt.Sprintf("This "+dataSourceLabel+" block has the "+
						"same data type and name as a previous block declared "+
						"at %s. Each "+dataSourceLabel+" must have a unique name per builder type.",
						existing.block.DefRange.Ptr()),
					Subject: datasource.block.DefRange.Ptr(),
				})
				continue
			}
			if cfg.Datasources == nil {
				cfg.Datasources = Datasources{}
			}
			cfg.Datasources[ref] = *datasource
		}
	}

	return diags
}
