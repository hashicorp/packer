// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
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
	// Channels
	Channels []string

	HCL2Ref
}

var bucketNameRegexp = regexp.MustCompile("^[a-zA-Z0-9-]{3,36}$")

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
		Channels     []string          `hcl:"channels,optional"`
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

	// No need to check the bucket name here if it's empty, since it can
	// be set through the `HCP_PACKER_BUCKET_NAME` environment var.
	//
	// If both are unset, creating the build on HCP Packer will fail, and
	// so will the packer build command.
	if b.Slug != "" && !bucketNameRegexp.MatchString(b.Slug) {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("%s.bucket_name can only contain between 3 and 36 ASCII letters, numbers and hyphens", buildHCPPackerRegistryLabel),
			Subject:  block.DefRange.Ptr(),
		})
	}

	par.Slug = b.Slug
	par.Description = b.Description
	par.Channels = b.Channels

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

// ExtractBuildProvisionerHCL extracts all provisioner blocks from the build
// blocks in the configuration and returns them as raw HCL content.
// This is used to publish provisioner configurations as enforced blocks
// to HCP Packer, so that other builds against the same bucket will
// automatically have these provisioners injected.
func (cfg *PackerConfig) ExtractBuildProvisionerHCL() (string, error) {
	sourceFiles := cfg.parser.Files()

	var buf strings.Builder

	for filename, file := range sourceFiles {
		// hclwrite only supports HCL native syntax, skip JSON and variable files
		if !strings.HasSuffix(filename, hcl2FileExt) {
			continue
		}

		wf, diags := hclwrite.ParseConfig(file.Bytes, filename, hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			continue
		}

		for _, block := range wf.Body().Blocks() {
			if block.Type() != buildLabel {
				continue
			}

			for _, inner := range block.Body().Blocks() {
				if inner.Type() != buildProvisionerLabel {
					continue
				}

				buf.Write(inner.BuildTokens(nil).Bytes())
				buf.WriteString("\n")
			}
		}
	}

	return strings.TrimSpace(buf.String()), nil
}
