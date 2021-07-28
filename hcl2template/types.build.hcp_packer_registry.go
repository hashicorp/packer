package hcl2template

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
)

type HCPPackerRegistryBlock struct {
	Description string
	Labels      map[string]string

	HCL2Ref HCL2Ref
}

func (b *HCPPackerRegistryBlock) WriteBucketConfig(bucket *packerregistry.Bucket) {
	if b == nil {
		return
	}
	bucket.Description = b.Description
	bucket.Labels = b.Labels
}

func (p *Parser) decodeHCPRegistry(block *hcl.Block) (*HCPPackerRegistryBlock, hcl.Diagnostics) {
	par := &HCPPackerRegistryBlock{}
	body := block.Body

	var b struct {
		Description string            `hcl:"description,optional"`
		Labels      map[string]string `hcl:"labels,optional"`
		Config      hcl.Body          `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(body, nil, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	if len(b.Description) > 255 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf(buildHCPPackerRegistryLabel + ".description should have a maximum length of 255 characters"),
			Subject:  block.DefRange.Ptr(),
		})
		return nil, diags
	}

	par.Description = b.Description
	par.Labels = b.Labels

	return par, diags
}
