package hcl2template

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
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

func (p *Parser) StartProvisioner(pb *ProvisionerBlock) (packer.Provisioner, hcl.Diagnostics) {
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
	flatProvisinerCfg, moreDiags := decodeHCL2Spec(pb.block, nil, provisioner)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	err = provisioner.Prepare(flatProvisinerCfg)
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
