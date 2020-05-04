package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/packer/packer"
)

// ProvisionerBlock references a detected but unparsed post processor
type PostProcessorBlock struct {
	PType string
	PName string

	HCL2Ref
}

func (p *PostProcessorBlock) String() string {
	return fmt.Sprintf(buildPostProcessorLabel+"-block %q %q", p.PType, p.PName)
}

func (p *Parser) decodePostProcessor(block *hcl.Block) (*PostProcessorBlock, hcl.Diagnostics) {
	var b struct {
		Name string   `hcl:"name,optional"`
		Rest hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	postProcessor := &PostProcessorBlock{
		PType:   block.Labels[0],
		PName:   b.Name,
		HCL2Ref: newHCL2Ref(block, b.Rest),
	}

	if !p.PostProcessorsSchemas.Has(postProcessor.PType) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  fmt.Sprintf("Unknown "+buildPostProcessorLabel+" type %q", postProcessor.PType),
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known "+buildPostProcessorLabel+"s: %v", p.PostProcessorsSchemas.List()),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}

	return postProcessor, diags
}

func (p *Parser) startPostProcessor(source *SourceBlock, pp *PostProcessorBlock, ectx *hcl.EvalContext, generatedVars map[string]string) (packer.PostProcessor, hcl.Diagnostics) {
	// ProvisionerBlock represents a detected but unparsed provisioner
	var diags hcl.Diagnostics

	postProcessor, err := p.PostProcessorsSchemas.Start(pp.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: fmt.Sprintf("Failed loading %s", pp.PType),
			Subject: pp.DefRange.Ptr(),
			Detail:  err.Error(),
		})
		return nil, diags
	}
	flatProvisinerCfg, moreDiags := decodeHCL2Spec(pp.Rest, ectx, postProcessor)
	diags = append(diags, moreDiags...)
	err = postProcessor.Configure(source.builderVariables(), flatProvisinerCfg, generatedVars)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed preparing %s", pp),
			Detail:   err.Error(),
			Subject:  pp.DefRange.Ptr(),
		})
		return nil, diags
	}
	return postProcessor, diags
}
