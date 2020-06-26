package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
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

func (cfg *PackerConfig) startPostProcessor(source SourceBlock, pp *PostProcessorBlock, ectx *hcl.EvalContext) (packer.PostProcessor, hcl.Diagnostics) {
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
	flatProvisinerCfg, moreDiags := decodeHCL2Spec(pp.Rest, ectx, postProcessor)
	diags = append(diags, moreDiags...)
	err = postProcessor.Configure(source.builderVariables(), flatProvisinerCfg)
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

type postProcessorsPrepare struct {
	cfg                *PackerConfig
	postProcessorBlock []*PostProcessorBlock
	src                SourceRef
}

// HCL2PostProcessorsPrepare is used by the CoreBuild at the runtime, after running the build and before running the post-processors,
// to interpolate any build variable by decoding and preparing it.
func (pp *postProcessorsPrepare) HCL2PostProcessorsPrepare(builderArtifact packer.Artifact) ([]packer.CoreBuildPostProcessor, hcl.Diagnostics) {
	src := pp.cfg.Sources[pp.src.Ref()]

	generatedData := make(map[string]interface{})
	if builderArtifact != nil {
		artifactStateData := builderArtifact.State("generated_data")
		if artifactStateData != nil {
			for k, v := range artifactStateData.(map[interface{}]interface{}) {
				generatedData[k.(string)] = v
			}
		}
	}

	variables := make(Variables)
	for k, v := range generatedData {
		if value, ok := v.(string); ok {
			variables[k] = &Variable{
				DefaultValue: cty.StringVal(value),
				Type:         cty.String,
			}
		}
	}
	variablesVal, _ := variables.Values()

	generatedVariables := map[string]cty.Value{
		sourcesAccessor: cty.ObjectVal(map[string]cty.Value{
			"type": cty.StringVal(src.Type),
			"name": cty.StringVal(src.Name),
		}),
		buildAccessor: cty.ObjectVal(variablesVal),
	}

	return pp.cfg.getCoreBuildPostProcessors(src, pp.postProcessorBlock, pp.cfg.EvalContext(generatedVariables))
}
