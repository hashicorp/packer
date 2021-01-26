package hcl2template

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	hcl2shim "github.com/hashicorp/packer/hcl2template/shim"
	"github.com/zclconf/go-cty/cty"
)

// HCL2Provisioner has a reference to the part of the HCL2 body where it is
// defined, allowing to completely reconfigure the Provisioner right before
// calling Provision: with contextual variables.
// This permits using "${build.ID}" values for example.
type HCL2Provisioner struct {
	Provisioner      packersdk.Provisioner
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

	flatProvisionerCfg, moreDiags := decodeHCL2Spec(p.provisionerBlock.HCL2Ref.Rest, ectx, p.Provisioner)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return diags
	}

	// In case of cty.Unknown values, this will write a equivalent placeholder of the same type
	// Unknown types are not recognized by the json marshal during the RPC call and we have to do this here
	// to avoid json parsing failures when running the validate command.
	// We don't do this before so we can validate if variable types matches correctly on decodeHCL2Spec.
	flatProvisionerCfg = hcl2shim.WriteUnknownPlaceholderValues(flatProvisionerCfg)

	return p.Provisioner.Prepare(p.builderVariables, flatProvisionerCfg, p.override)
}

func (p *HCL2Provisioner) Prepare(args ...interface{}) error {
	return p.Provisioner.Prepare(args...)
}

func (p *HCL2Provisioner) Provision(ctx context.Context, ui packersdk.Ui, c packersdk.Communicator, vars map[string]interface{}) error {
	err := p.HCL2Prepare(vars)
	if err != nil {
		return err
	}
	return p.Provisioner.Provision(ctx, ui, c, vars)
}
