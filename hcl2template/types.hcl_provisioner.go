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

// HCL2Provisioner has a reference to the part of the HCL2 body where it is
// defined, allowing to completely reconfigure the Provisioner right before
// calling Provision: with contextual variables.
// This permits using "${build.ID}" values for example.
type HCL2Provisioner struct {
	Provisioner      packer.Provisioner
	provisionerBlock *ProvisionerBlock
	evalContext      *hcl.EvalContext
	builderVariables map[string]string
	override         map[string]interface{}
}

func (p *HCL2Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.Provisioner.ConfigSpec()
}

func (p *HCL2Provisioner) HCL2Prepare(buildVars map[string]interface{}) error {
	var diags hcl.Diagnostics
	ectx := p.evalContext
	if len(buildVars) > 0 {
		ectx = p.evalContext.NewChild()
		buildValues := map[string]cty.Value{}
		if !p.evalContext.Variables[buildAccessor].IsNull() {
			buildValues = p.evalContext.Variables[buildAccessor].AsValueMap()
		}
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

	flatProvisionerCfg, moreDiags := decodeHCL2Spec(p.provisionerBlock.HCL2Ref.Rest, ectx, p.Provisioner)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return diags
	}
	return p.Provisioner.Prepare(p.builderVariables, flatProvisionerCfg, p.override)
}

func (p *HCL2Provisioner) Prepare(args ...interface{}) error {
	return p.Provisioner.Prepare(args...)
}

func (p *HCL2Provisioner) Provision(ctx context.Context, ui packersdk.Ui, c packer.Communicator, vars map[string]interface{}) error {
	err := p.HCL2Prepare(vars)
	if err != nil {
		return err
	}
	return p.Provisioner.Provision(ctx, ui, c, vars)
}
