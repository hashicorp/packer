package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
)

// PackerConfig represents a loaded packer config
type PackerConfig struct {
	Sources map[SourceRef]*Source

	Variables PackerV1Variables

	Builds Builds
}

func (p *Parser) CoreBuildProvisioners(blocks []*ProvisionerBlock) ([]packer.CoreBuildProvisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildProvisioner{}
	for _, pb := range blocks {
		provisioner, moreDiags := p.StartProvisioner(pb)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		res = append(res, packer.CoreBuildProvisioner{
			PType:       pb.PType,
			Provisioner: provisioner,
		})
	}
	return res, diags
}

func (p *Parser) CoreBuildPostProcessors(blocks []*PostProcessorBlock) ([]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := []packer.CoreBuildPostProcessor{}
	for _, pp := range blocks {
		postProcessor, moreDiags := p.StartPostProcessor(pp)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		res = append(res, packer.CoreBuildPostProcessor{
			PostProcessor: postProcessor,
			PType:         pp.PType,
		})
	}
	return res, diags
}

func (p *Parser) getBuilds(cfg *PackerConfig) ([]packer.Build, hcl.Diagnostics) {
	res := []packer.Build{}
	var diags hcl.Diagnostics

	for _, build := range cfg.Builds {
		for _, from := range build.Froms {
			src, found := cfg.Sources[from]
			if !found {
				diags = append(diags, &hcl.Diagnostic{
					Summary:  "Unknown " + sourceLabel + " " + from.String(),
					Subject:  build.HCL2Ref.DeclRange.Ptr(),
					Severity: hcl.DiagError,
				})
				continue
			}
			builder, moreDiags := p.StartBuilder(src)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			provisioners, moreDiags := p.CoreBuildProvisioners(build.ProvisionerBlocks)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			postProcessors, moreDiags := p.CoreBuildPostProcessors(build.PostProcessors)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			pcb := &packer.CoreBuild{
				Type:           src.Type,
				Builder:        builder,
				Provisioners:   provisioners,
				PostProcessors: [][]packer.CoreBuildPostProcessor{postProcessors},
				Variables:      cfg.Variables,
			}
			res = append(res, pcb)
		}
	}
	return res, diags
}

func (p *Parser) Parse(path string) ([]packer.Build, hcl.Diagnostics) {
	cfg, diags := p.parse(path)
	if diags.HasErrors() {
		return nil, diags
	}

	return p.getBuilds(cfg)
}
