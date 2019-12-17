package hcl2template

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/packer"
)

// ProvisionerBlock represents a parsed provisioner
type ProvisionerBlock struct {
	PType string
	block *hcl.Block
}

func (p *Parser) decodeProvisioner(block *hcl.Block) (*ProvisionerBlock, hcl.Diagnostics) {
	provisioner := &ProvisionerBlock{
		PType: block.Labels[0],
		block: block,
	}
	var diags hcl.Diagnostics

	if !p.ProvisionersSchemas.Has(provisioner.PType) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unknown " + buildProvisionerLabel + " type " + provisioner.PType,
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known provisioners: %v", p.ProvisionersSchemas.List()),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}
	return provisioner, diags
}

func (p *Parser) StartProvisioner(pb *ProvisionerBlock, generatedVars []string) (packer.Provisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	provisioner, err := p.ProvisionersSchemas.Start(pb.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Failed loading " + pb.block.Type,
			Subject: pb.block.LabelRanges[0].Ptr(),
			Detail:  err.Error(),
		})
		return nil, diags
	}
	flatProvisionerCfg, moreDiags := decodeHCL2Spec(pb.block, nil, provisioner)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	// manipulate generatedVars from builder to add to the interfaces being
	// passed to the provisioner Prepare()

	// If the builder has provided a list of to-be-generated variables that
	// should be made accessible to provisioners, pass that list into
	// the provisioner prepare() so that the provisioner can appropriately
	// validate user input against what will become available. Otherwise,
	// only pass the default variables, using the basic placeholder data.
	generatedPlaceholderMap := packer.BasicPlaceholderData()
	if generatedVars != nil {
		for _, k := range generatedVars {
			generatedPlaceholderMap[k] = fmt.Sprintf("Generated_%s. "+
				common.PlaceholderMsg, k)
		}
	}
	// configs := make([]interface{}, 2)
	// configs = append(, flatProvisionerCfg)
	// configs = append(configs, generatedPlaceholderMap)
	err = provisioner.Prepare(flatProvisionerCfg, generatedPlaceholderMap)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Failed preparing " + pb.block.Type,
			Detail:   err.Error(),
			Subject:  pb.block.DefRange.Ptr(),
		})
		return nil, diags
	}
	return provisioner, diags
}
