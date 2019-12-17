package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
)

// PostProcessorBlock represents a parsed PostProcessorBlock
type PostProcessorBlock struct {
	PType string
	block *hcl.Block
}

func (p *Parser) decodePostProcessorGroup(block *hcl.Block) (*PostProcessorBlock, hcl.Diagnostics) {
	postProcessor := &PostProcessorBlock{
		PType: block.Labels[0],
		block: block,
	}
	var diags hcl.Diagnostics

	if !p.PostProcessorsSchemas.Has(postProcessor.PType) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + buildPostProcessorLabel + " type " + postProcessor.PType,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known provisioners: %v", p.ProvisionersSchemas.List()),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}

	return postProcessor, diags
}

func (p *Parser) StartPostProcessor(pp *PostProcessorBlock) (packer.PostProcessor, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	postProcessor, err := p.PostProcessorsSchemas.Start(pp.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Failed loading " + pp.block.Type,
			Subject: pp.block.LabelRanges[0].Ptr(),
			Detail:  err.Error(),
		})
		return nil, diags
	}
	flatProvisinerCfg, moreDiags := decodeHCL2Spec(pp.block, nil, postProcessor)
	diags = append(diags, moreDiags...)
	err = postProcessor.Configure(flatProvisinerCfg)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed preparing " + pp.block.Type,
			Detail:   err.Error(),
			Subject:  pp.block.DefRange.Ptr(),
		})
		return nil, diags
	}
	return postProcessor, diags
}
