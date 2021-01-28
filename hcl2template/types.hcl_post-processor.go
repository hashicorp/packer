package hcl2template

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/zclconf/go-cty/cty"
)

// HCL2PostProcessor has a reference to the part of the HCL2 body where it is
// defined, allowing to completely reconfigure the PostProcessor right before
// calling PostProcess: with contextual variables.
// This permits using "${build.ID}" values for example.
type HCL2PostProcessor struct {
	PostProcessor      packersdk.PostProcessor
	postProcessorBlock *PostProcessorBlock
	evalContext        *hcl.EvalContext
	builderVariables   map[string]string
}

func (p *HCL2PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return p.PostProcessor.ConfigSpec()
}

func (p *HCL2PostProcessor) HCL2Prepare(buildVars map[string]interface{}) error {
	var diags hcl.Diagnostics
	ectx := p.evalContext
	if len(buildVars) > 0 {
		ectx = p.evalContext.NewChild()
		buildValues := map[string]cty.Value{}
		for k, v := range buildVars {
			val, err := ConvertPluginConfigValueToHCLValue(v)
			if err != nil {
				return err
			}

			buildValues[k] = val
		}
		ectx.Variables = map[string]cty.Value{
			buildAccessor: cty.ObjectVal(buildValues),
		}
	}

	flatPostProcessorCfg, moreDiags := decodeHCL2Spec(p.postProcessorBlock.HCL2Ref.Rest, ectx, p.PostProcessor)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return diags
	}

	// In case of cty.Unknown values, this will write a equivalent placeholder of the same type
	// Unknown types are not recognized by the json marshal during the RPC call and we have to do this here
	// to avoid json parsing failures when running the validate command.
	// We don't do this before so we can validate if variable types matches correctly on decodeHCL2Spec.
	flatPostProcessorCfg = hcl2shim.WriteUnknownPlaceholderValues(flatPostProcessorCfg)

	return p.PostProcessor.Configure(p.builderVariables, flatPostProcessorCfg)
}

func (p *HCL2PostProcessor) Configure(args ...interface{}) error {
	return p.PostProcessor.Configure(args...)
}

func (p *HCL2PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	generatedData := make(map[string]interface{})
	if artifactStateData, ok := artifact.State("generated_data").(map[interface{}]interface{}); ok {
		for k, v := range artifactStateData {
			generatedData[k.(string)] = v
		}
	}

	err := p.HCL2Prepare(generatedData)
	if err != nil {
		return nil, false, false, err
	}
	return p.PostProcessor.PostProcess(ctx, ui, artifact)
}
