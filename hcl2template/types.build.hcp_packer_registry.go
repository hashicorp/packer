package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
)

type HCPPackerRegistryBlock struct {
	Description string

	HCL2Ref HCL2Ref
}

func (b *HCPPackerRegistryBlock) WriteBucket(bucket *packerregistry.Bucket) {
	if b == nil {
		return
	}
	bucket.Description = b.Description
}

func (p *Parser) decodeHCPRegistry(block *hcl.Block) (*HCPPackerRegistryBlock, hcl.Diagnostics) {
	par := &HCPPackerRegistryBlock{}
	body := block.Body

	var b struct {
		Description string   `hcl:"description,optional"`
		Config      hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(body, nil, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	par.Description = b.Description

	return par, diags
}
