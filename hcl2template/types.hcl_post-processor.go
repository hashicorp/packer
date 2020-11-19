package hcl2template

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/zclconf/go-cty/cty"
)

// HCL2PostProcessor has a reference to the part of the HCL2 body where it is
// defined, allowing to completely reconfigure the PostProcessor right before
// calling PostProcess: with contextual variables.
// This permits using "${build.ID}" values for example.
type HCL2PostProcessor struct {
	PostProcessor      packer.PostProcessor
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
			switch v := v.(type) {
			case string:
				buildValues[k] = cty.StringVal(v)
			case int64:
				buildValues[k] = cty.NumberIntVal(v)
			case uint64:
				buildValues[k] = cty.NumberUIntVal(v)
			case bool:
				buildValues[k] = cty.BoolVal(v)
			default:
				return fmt.Errorf("unhandled buildvar type: %T", v)
			}
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
	return p.PostProcessor.Configure(p.builderVariables, flatPostProcessorCfg)
}

func (p *HCL2PostProcessor) Configure(args ...interface{}) error {
	return p.PostProcessor.Configure(args...)
}

func (p *HCL2PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
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
