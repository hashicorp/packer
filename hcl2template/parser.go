package hcl2template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/dynblock"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer/packer"
)

const (
	sourceLabel       = "source"
	variablesLabel    = "variables"
	variableLabel     = "variable"
	localsLabel       = "locals"
	buildLabel        = "build"
	communicatorLabel = "communicator"
)

var configSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: sourceLabel, LabelNames: []string{"type", "name"}},
		{Type: variablesLabel},
		{Type: variableLabel, LabelNames: []string{"name"}},
		{Type: localsLabel},
		{Type: buildLabel},
		{Type: communicatorLabel, LabelNames: []string{"type", "name"}},
	},
}

// Parser helps you parse HCL folders. It will parse an hcl file or directory
// and start builders, provisioners and post-processors to configure them with
// the parsed HCL and then return a []packer.Build. Packer will use that list
// of Builds to run everything in order.
type Parser struct {
	*hclparse.Parser

	BuilderSchemas packer.BuilderStore

	ProvisionersSchemas packer.ProvisionerStore

	PostProcessorsSchemas packer.PostProcessorStore
}

const (
	hcl2FileExt        = ".pkr.hcl"
	hcl2JsonFileExt    = ".pkr.json"
	hcl2VarFileExt     = ".auto.pkrvars.hcl"
	hcl2VarJsonFileExt = ".auto.pkrvars.json"
)

func (p *Parser) parse(filename string, varFiles []string, argVars map[string]string) (*PackerConfig, hcl.Diagnostics) {

	var files []*hcl.File
	var diags hcl.Diagnostics

	// parse config files
	{
		hclFiles, jsonFiles, moreDiags := GetHCL2Files(filename, hcl2FileExt, hcl2JsonFileExt)
		diags = append(diags, moreDiags...)
		if len(hclFiles)+len(jsonFiles) == 0 {
			diags = append(moreDiags, &hcl.Diagnostic{
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
	cfg := &PackerConfig{
		Basedir: basedir,
	}

	// Decode variable blocks so that they are available later on. Here locals
	// can use input variables so we decode them firsthand.
	var locals []*Local
	{
		for _, file := range files {
			diags = append(diags, cfg.decodeInputVariables(file)...)
		}

		for _, file := range files {
			moreLocals, morediags := cfg.parseLocalVariables(file)
			diags = append(diags, morediags...)
			locals = append(locals, moreLocals...)
		}
	}

	// parse var files
	{
		hclVarFiles, jsonVarFiles, moreDiags := GetHCL2Files(filename, hcl2VarFileExt, hcl2VarJsonFileExt)
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

	_, moreDiags := cfg.InputVariables.Values()
	diags = append(diags, moreDiags...)
	_, moreDiags = cfg.LocalVariables.Values()
	diags = append(diags, moreDiags...)
	diags = append(diags, cfg.evaluateLocalVariables(locals)...)

	// decode the actual content
	for _, file := range files {
		diags = append(diags, p.decodeConfig(file, cfg)...)
	}

	return cfg, diags
}

// decodeConfig looks in the found blocks for everything that is not a variable
// block. It should be called after parsing input variables and locals so that
// they can be referenced.
func (p *Parser) decodeConfig(f *hcl.File, cfg *PackerConfig) hcl.Diagnostics {
	var diags hcl.Diagnostics

	body := dynblock.Expand(f.Body, cfg.EvalContext(nil))
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
			if existing := cfg.Sources[ref]; existing != nil {
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
				cfg.Sources = map[SourceRef]*SourceBlock{}
			}
			cfg.Sources[ref] = source

		case buildLabel:
			build, moreDiags := p.decodeBuildConfig(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			cfg.Builds = append(cfg.Builds, build)

		}
	}

	return diags
}
