package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// ProvisionerBlock references a detected but unparsed post processor
type PostProcessorBlock struct {
	PType             string
	PName             string
	OnlyExcept        OnlyExcept
	KeepInputArtifact *bool

	HCL2Ref
}

func (p *PostProcessorBlock) String() string {
	return fmt.Sprintf(buildPostProcessorLabel+"-block %q %q", p.PType, p.PName)
}

func (p *Parser) decodePostProcessor(block *hcl.Block) (*PostProcessorBlock, hcl.Diagnostics) {
	var b struct {
		Name              string   `hcl:"name,optional"`
		Only              []string `hcl:"only,optional"`
		Except            []string `hcl:"except,optional"`
		KeepInputArtifact *bool    `hcl:"keep_input_artifact,optional"`
		Rest              hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	postProcessor := &PostProcessorBlock{
		PType:             block.Labels[0],
		PName:             b.Name,
		OnlyExcept:        OnlyExcept{Only: b.Only, Except: b.Except},
		HCL2Ref:           newHCL2Ref(block, b.Rest),
		KeepInputArtifact: b.KeepInputArtifact,
	}

	diags = diags.Extend(postProcessor.OnlyExcept.Validate())
	if diags.HasErrors() {
		return nil, diags
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

func (cfg *PackerConfig) startPostProcessor(source SourceBlock, pp *PostProcessorBlock, ectx *hcl.EvalContext) (packersdk.PostProcessor, hcl.Diagnostics) {
	// ProvisionerBlock represents a detected but unparsed provisioner
	var diags hcl.Diagnostics

	postProcessor, err := cfg.postProcessorsSchemas.Start(pp.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: fmt.Sprintf("Failed loading %s", pp.PType),
			Subject: pp.DefRange.Ptr(),
			Detail:  err.Error(),
		})
		return nil, diags
	}
	hclPostProcessor := &HCL2PostProcessor{
		PostProcessor:      postProcessor,
		postProcessorBlock: pp,
		evalContext:        ectx,
		builderVariables:   source.builderVariables(),
	}
	err = hclPostProcessor.HCL2Prepare(nil)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed preparing %s", pp),
			Detail:   err.Error(),
			Subject:  pp.DefRange.Ptr(),
		})
		return nil, diags
	}
	return hclPostProcessor, diags
}
