package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

type HCPPackerRegistryBlock struct {
	// Bucket slug
	Slug string
	// Bucket description
	Description string
	// Bucket labels
	BucketLabels map[string]string
	// Build labels
	BuildLabels map[string]string

	HCL2Ref
}

func (p *Parser) decodeHCPRegistry(block *hcl.Block, cfg *PackerConfig) (*HCPPackerRegistryBlock, hcl.Diagnostics) {
	par := &HCPPackerRegistryBlock{}
	body := block.Body

	var b struct {
		Slug        string `hcl:"bucket_name,optional"`
		Description string `hcl:"description,optional"`
		//Deprecated labels for bucket_labels
		Labels       map[string]string `hcl:"labels,optional"`
		BucketLabels map[string]string `hcl:"bucket_labels,optional"`
		BuildLabels  map[string]string `hcl:"build_labels,optional"`
		Config       hcl.Body          `hcl:",remain"`
	}
	ectx := cfg.EvalContext(BuildContext, nil)
	diags := gohcl.DecodeBody(body, ectx, &b)
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

	par.Slug = b.Slug
	par.Description = b.Description

	if len(b.Labels) > 0 && len(b.BucketLabels) > 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s.labels and %[1]s.bucket_labels are mutually exclusive; please use the recommended argument %[1]s.bucket_labels", buildHCPPackerRegistryLabel),
			Subject:  block.DefRange.Ptr(),
		})
		return nil, diags
	}

	if len(b.Labels) > 0 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  fmt.Sprintf("the argument %s.labels has been deprecated and will be removed in the next minor release; please use %[1]s.bucket_labels", buildHCPPackerRegistryLabel),
		})

		b.BucketLabels = b.Labels
	}

	par.BucketLabels = b.BucketLabels
	par.BuildLabels = b.BuildLabels

	return par, diags
}
