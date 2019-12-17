package hcl2template

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/packer/packer"
)

const (
	sourceLabel       = "source"
	variablesLabel    = "variables"
	buildLabel        = "build"
	communicatorLabel = "communicator"
)

var configSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: sourceLabel, LabelNames: []string{"type", "name"}},
		{Type: variablesLabel},
		{Type: buildLabel},
		{Type: communicatorLabel, LabelNames: []string{"type", "name"}},
	},
}

type Parser struct {
	*hclparse.Parser

	BuilderSchemas packer.BuilderStore

	ProvisionersSchemas packer.ProvisionerStore

	PostProcessorsSchemas packer.PostProcessorStore
}

const hcl2FileExt = ".pkr.hcl"

func (p *Parser) parse(filename string) (*PackerConfig, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	hclFiles := []string{}
	jsonFiles := []string{}
	if strings.HasSuffix(filename, hcl2FileExt) {
		hclFiles = append(hclFiles, filename)
	} else if strings.HasSuffix(filename, ".json") {
		jsonFiles = append(jsonFiles, filename)
	} else {
		fileInfos, err := ioutil.ReadDir(filename)
		if err != nil {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Cannot read hcl directory",
				Detail:   err.Error(),
			}
			diags = append(diags, diag)
		}
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				continue
			}
			filename := filepath.Join(filename, fileInfo.Name())
			if strings.HasSuffix(filename, hcl2FileExt) {
				hclFiles = append(hclFiles, filename)
			} else if strings.HasSuffix(filename, ".json") {
				jsonFiles = append(jsonFiles, filename)
			}
		}
	}

	var files []*hcl.File
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

	cfg := &PackerConfig{}
	for _, file := range files {
		moreDiags := p.parseFile(file, cfg)
		diags = append(diags, moreDiags...)
	}

	return cfg, diags
}

// parseFile filename content into cfg.
//
// parseFile may be called multiple times with the same cfg on a different file.
//
// parseFile returns as complete a config as we can manage, even if there are
// errors, since a partial result can be useful for careful analysis by
// development tools such as text editor extensions.
func (p *Parser) parseFile(f *hcl.File, cfg *PackerConfig) hcl.Diagnostics {
	var diags hcl.Diagnostics

	content, moreDiags := f.Body.Content(configSchema)
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
				cfg.Sources = map[SourceRef]*Source{}
			}
			cfg.Sources[ref] = source

		case variablesLabel:
			if cfg.Variables == nil {
				cfg.Variables = PackerV1Variables{}
			}

			moreDiags := cfg.Variables.decodeConfig(block)
			if moreDiags.HasErrors() {
				continue
			}
			diags = append(diags, moreDiags...)

		case buildLabel:
			build, moreDiags := p.decodeBuildConfig(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			cfg.Builds = append(cfg.Builds, build)

		default:
			panic(fmt.Sprintf("unexpected block type %q", block.Type)) // TODO(azr): err
		}
	}

	return diags
}
