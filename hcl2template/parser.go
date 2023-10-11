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
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	pkrfunction "github.com/hashicorp/packer/hcl2template/function"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
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
	files := map[string]*hcl.File{}
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
			files[filename] = f
		}
		for _, filename := range jsonFiles {
			f, moreDiags := p.ParseJSONFile(filename)
			diags = append(diags, moreDiags...)
			files[filename] = f
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
	versionDiags := cfg.CheckCoreVersionRequirements(p.CorePackerVersion)
	diags = append(diags, versionDiags...)
	if versionDiags.HasErrors() {
		return cfg, diags
	}

	for _, file := range files {
		diags = append(diags, cfg.decodeFile(file)...)
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

// decodeFile attempts to decode the configuration from a HCL file, and starts
// populating the config from it.
func (cfg *PackerConfig) decodeFile(file *hcl.File) hcl.Diagnostics {
	var diags hcl.Diagnostics

	content, moreDiags := file.Body.Content(configSchema)
	diags = append(diags, moreDiags...)
	// If basic parsing failed, we should not continue
	if diags.HasErrors() {
		return diags
	}

	for _, block := range content.Blocks {
		diags = append(diags, cfg.decodeBlock(block)...)
	}

	return diags
}

func (cfg *PackerConfig) decodeBlock(block *hcl.Block) hcl.Diagnostics {
	switch block.Type {
	case packerLabel:
		return cfg.decodePackerBlock(block)
	case dataSourceLabel:
		return cfg.decodeDatasource(block)
	case variableLabel:
		return cfg.decodeVariableBlock(block)
	case variablesLabel:
		return cfg.decodeVariablesBlock(block)
	case localLabel:
		return cfg.decodeLocalBlock(block)
	case localsLabel:
		return cfg.decodeLocalsBlock(block)
	// NOTE: Both build and source blocks can be dynamically expanded.
	//
	// This means that while at this time we can already get some information
	// about them, we may not be able to decode their final form at this time
	// and we can only do so when their dependencies are evaluated.
	case buildLabel:
		return cfg.decodeBuildBlock(block)
	case sourceLabel:
		return cfg.decodeSourceBlock(block)
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid block type",
			Detail:   fmt.Sprintf("The block %q is not a valid top-level block", block.Type),
			Subject:  &block.DefRange,
		},
	}
}

func (cfg *PackerConfig) decodePackerBlock(block *hcl.Block) hcl.Diagnostics {
	var diags hcl.Diagnostics
	content, contentDiags := block.Body.Content(packerBlockSchema)
	diags = append(diags, contentDiags...)

	// We ignore "packer_version"" here because
	// sniffCoreVersionRequirements already dealt with that
	for _, innerBlock := range content.Blocks {
		switch innerBlock.Type {
		case "required_plugins":
			reqs, reqsDiags := decodeRequiredPluginsBlock(innerBlock)
			diags = append(diags, reqsDiags...)
			cfg.Packer.RequiredPlugins = append(cfg.Packer.RequiredPlugins, reqs)
		default:
			continue
		}

	}

	return diags
}

func (cfg *PackerConfig) decodeDatasource(block *hcl.Block) hcl.Diagnostics {
	datasource, diags := cfg.decodeDataBlock(block)
	if diags.HasErrors() {
		return diags
	}
	ref := datasource.Ref()
	if existing, found := cfg.Datasources[ref]; found {
		return append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Duplicate " + dataSourceLabel + " block",
			Detail: fmt.Sprintf("This "+dataSourceLabel+" block has the "+
				"same data type and name as a previous block declared "+
				"at %s. Each "+dataSourceLabel+" must have a unique name per builder type.",
				existing.block.DefRange.Ptr()),
			Subject: datasource.block.DefRange.Ptr(),
		})
	}
	if cfg.Datasources == nil {
		cfg.Datasources = Datasources{}
	}
	cfg.Datasources[ref] = datasource

	datasource.getDependencies()

	return diags
}

func (cfg *PackerConfig) decodeDataBlock(block *hcl.Block) (*DatasourceBlock, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	r := &DatasourceBlock{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		block: block,
	}

	if !hclsyntax.ValidIdentifier(r.Type) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid data source name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}
	if !hclsyntax.ValidIdentifier(r.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid data resource name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[1],
		})
	}

	return r, diags
}

func (cfg *PackerConfig) decodeLocalBlock(block *hcl.Block) hcl.Diagnostics {
	name := block.Labels[0]

	content, diags := block.Body.Content(localBlockSchema)
	if !hclsyntax.ValidIdentifier(name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid local name",
			Detail:   badIdentifierDetail,
			Subject:  &block.LabelRanges[0],
		})
	}

	l := &LocalBlock{
		Name: name,
	}

	if attr, exists := content.Attributes["sensitive"]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &l.Sensitive)
		diags = append(diags, valDiags...)
	}

	if def, ok := content.Attributes["expression"]; ok {
		l.Expr = def.Expr
	}
	l.getDependencies()

	cfg.LocalBlocks = append(cfg.LocalBlocks, l)

	return diags
}

func (cfg *PackerConfig) decodeLocalsBlock(block *hcl.Block) hcl.Diagnostics {
	attrs, diags := block.Body.JustAttributes()

	for name, attr := range attrs {
		l := &LocalBlock{
			Name: name,
			Expr: attr.Expr,
		}
		l.getDependencies()
		cfg.LocalBlocks = append(cfg.LocalBlocks, l)
	}

	return diags
}

func (cfg *PackerConfig) decodeVariableBlock(block *hcl.Block) hcl.Diagnostics {
	// for input variables we allow to use env in the default value section.
	ectx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"env": pkrfunction.EnvFunc,
		},
	}

	return cfg.InputVariables.decodeVariableBlock(block, ectx)
}

func (cfg *PackerConfig) decodeVariablesBlock(block *hcl.Block) hcl.Diagnostics {
	// for input variables we allow to use env in the default value section.
	ectx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"env": pkrfunction.EnvFunc,
		},
	}

	var diags hcl.Diagnostics

	attrs, moreDiags := block.Body.JustAttributes()
	diags = append(diags, moreDiags...)
	for key, attr := range attrs {
		moreDiags = cfg.InputVariables.decodeVariable(key, attr, ectx)
		diags = append(diags, moreDiags...)
	}

	return diags
}

// decodeBuildBlock shallowly decodes a build block from the config.
//
// The final decoding step (which requires an up-to-date context) will be done
// when we need it.
func (cfg *PackerConfig) decodeBuildBlock(block *hcl.Block) hcl.Diagnostics {
	build := &BuildBlock{
		block: block,
	}

	if isDynamic(block) {
		build.dynamic = true
	}

	cfg.Builds = append(cfg.Builds, build)

	return nil
}

func (cfg *PackerConfig) decodeSourceBlock(block *hcl.Block) hcl.Diagnostics {
	source, diags := cfg.decodeSource(block)
	if diags.HasErrors() {
		return diags
	}

	if cfg.Sources == nil {
		cfg.Sources = map[SourceRef]SourceBlock{}
	}

	ref := source.Ref()
	if existing, found := cfg.Sources[ref]; found {
		return append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Duplicate " + sourceLabel + " block",
			Detail: fmt.Sprintf("This "+sourceLabel+" block has the "+
				"same builder type and name as a previous block declared "+
				"at %s. Each "+sourceLabel+" must have a unique name per builder type.",
				existing.block.DefRange.Ptr()),
			Subject: source.block.DefRange.Ptr(),
		})
	}

	cfg.Sources[ref] = source

	return diags
}

// decodeBuildSource reads a used source block from a build:
//
//	build {
//	  source "type.example" {
//	    name = "local_name"
//	  }
//	}
func (cfg *PackerConfig) decodeBuildSource(block *hcl.Block) (SourceUseBlock, hcl.Diagnostics) {
	ref := sourceRefFromString(block.Labels[0])
	out := SourceUseBlock{SourceRef: ref}
	var b struct {
		Name string   `hcl:"name,optional"`
		Rest hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)
	if diags.HasErrors() {
		return out, diags
	}
	out.LocalName = b.Name
	out.Body = b.Rest
	return out, nil
}

func (source *SourceBlock) finalizeDecodeSource(cfg *PackerConfig) hcl.Diagnostics {
	if source.Ready {
		return nil
	}

	source.Ready = true
	dyn := dynblock.Expand(source.block.Body, cfg.EvalContext(DatasourceContext, nil))
	// Expand without a base schema since nothing is known in advance for a
	// source, but we still want to expand dynamic blocks if any
	_, rem, diags := dyn.PartialContent(&hcl.BodySchema{})

	// Only try to expand once, regardless of whether the source succeeded
	// to expand dynamic data or not.
	source.Ready = true
	if diags.HasErrors() {
		return diags
	}

	source.block = &hcl.Block{
		Labels: []string{
			source.Type,
			source.Name,
		},
		Body: rem,
	}

	return diags
}

func (cfg *PackerConfig) decodeSource(block *hcl.Block) (SourceBlock, hcl.Diagnostics) {
	source := SourceBlock{
		Type:  block.Labels[0],
		Name:  block.Labels[1],
		block: block,
	}

	if isDynamic(block) {
		source.dynamic = true
	}

	var diags hcl.Diagnostics

	return source, diags
}
