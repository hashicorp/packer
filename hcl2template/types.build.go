// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

const (
	buildFromLabel = "from"

	buildSourceLabel = "source"

	buildProvisionerLabel = "provisioner"

	buildErrorCleanupProvisionerLabel = "error-cleanup-provisioner"

	buildPostProcessorLabel = "post-processor"

	buildPostProcessorsLabel = "post-processors"

	buildHCPPackerRegistryLabel = "hcp_packer_registry"
)

var buildSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: buildFromLabel, LabelNames: []string{"type"}},
		{Type: sourceLabel, LabelNames: []string{"reference"}},
		{Type: buildProvisionerLabel, LabelNames: []string{"type"}},
		{Type: buildErrorCleanupProvisionerLabel, LabelNames: []string{"type"}},
		{Type: buildPostProcessorLabel, LabelNames: []string{"type"}},
		{Type: buildPostProcessorsLabel, LabelNames: []string{}},
		{Type: buildHCPPackerRegistryLabel},
	},
}

var postProcessorsSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: buildPostProcessorLabel, LabelNames: []string{"type"}},
	},
}

// BuildBlock references an HCL 'build' block and it content, for example :
//
//	build {
//		sources = [
//			...
//		]
//		provisioner "" { ... }
//		post-processor "" { ... }
//	}
type BuildBlock struct {
	// Name is a string representing the named build to show in the logs
	Name string

	// A description of what this build does, it could be used in a inspect
	// call for example.
	Description string

	// HCPPackerRegistry contains the configuration for publishing the image to the HCP Packer Registry.
	HCPPackerRegistry *HCPPackerRegistryBlock

	// Sources is the list of sources that we want to start in this build block.
	Sources []SourceUseBlock

	// ProvisionerBlocks references a list of HCL provisioner block that will
	// will be ran against the sources.
	ProvisionerBlocks []*ProvisionerBlock

	// ErrorCleanupProvisionerBlock references a special provisioner block that
	// will be ran only if the provision step fails.
	ErrorCleanupProvisionerBlock *ProvisionerBlock

	// PostProcessorLists references the lists of lists of HCL post-processors
	// block that will be run against the artifacts from the provisioning
	// steps.
	PostProcessorsLists [][]*PostProcessorBlock

	HCL2Ref HCL2Ref
}

type Builds []*BuildBlock

// decodeBuildConfig is called when a 'build' block has been detected. It will
// load the references to the contents of the build block.
func (p *Parser) decodeBuildConfig(block *hcl.Block, cfg *PackerConfig) (*BuildBlock, hcl.Diagnostics) {
	var b struct {
		Name        string   `hcl:"name,optional"`
		Description string   `hcl:"description,optional"`
		FromSources []string `hcl:"sources,optional"`
		Config      hcl.Body `hcl:",remain"`
	}

	body := block.Body
	diags := gohcl.DecodeBody(body, cfg.EvalContext(LocalContext, nil), &b)
	if diags.HasErrors() {
		return nil, diags
	}

	build := &BuildBlock{
		HCL2Ref: newHCL2Ref(block, b.Config),
	}

	build.Name = b.Name
	build.Description = b.Description
	build.HCL2Ref.DefRange = block.DefRange

	// Expose build.name during parsing of pps and provisioners
	ectx := cfg.EvalContext(BuildContext, nil)
	ectx.Variables[buildAccessor] = cty.ObjectVal(map[string]cty.Value{
		"name": cty.StringVal(b.Name),
	})

	// We rely on `hadSource` to determine which error to proc.
	//
	// If a source block is referenced in the build block, but isn't valid, we
	// cannot rely on the `build.Sources' since it's only populated when a valid
	// source is processed.
	hadSource := false

	for _, buildFrom := range b.FromSources {
		hadSource = true

		ref := sourceRefFromString(buildFrom)

		if ref == NoSource ||
			!hclsyntax.ValidIdentifier(ref.Type) ||
			!hclsyntax.ValidIdentifier(ref.Name) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid " + sourceLabel + " reference",
				Detail: "A " + sourceLabel + " type is made of three parts that are" +
					"split by a dot `.`; each part must start with a letter and " +
					"may contain only letters, digits, underscores, and dashes." +
					"A valid source reference looks like: `source.type.name`",
				Subject: block.DefRange.Ptr(),
			})
			continue
		}

		// source with no body
		build.Sources = append(build.Sources, SourceUseBlock{SourceRef: ref})
	}

	body = b.Config
	content, moreDiags := body.Content(buildSchema)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	for _, block := range content.Blocks {
		switch block.Type {
		case buildHCPPackerRegistryLabel:
			if build.HCPPackerRegistry != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Only one " + buildHCPPackerRegistryLabel + " is allowed"),
					Subject:  block.DefRange.Ptr(),
				})
				continue
			}
			hcpPackerRegistry, moreDiags := p.decodeHCPRegistry(block, cfg)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.HCPPackerRegistry = hcpPackerRegistry
		case sourceLabel:
			hadSource = true
			ref, moreDiags := p.decodeBuildSource(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.Sources = append(build.Sources, ref)
		case buildProvisionerLabel:
			p, moreDiags := p.decodeProvisioner(block, ectx)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.ProvisionerBlocks = append(build.ProvisionerBlocks, p)
		case buildErrorCleanupProvisionerLabel:
			if build.ErrorCleanupProvisionerBlock != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Only one " + buildErrorCleanupProvisionerLabel + " is allowed"),
					Subject:  block.DefRange.Ptr(),
				})
				continue
			}
			p, moreDiags := p.decodeProvisioner(block, ectx)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.ErrorCleanupProvisionerBlock = p
		case buildPostProcessorLabel:
			pp, moreDiags := p.decodePostProcessor(block, ectx)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.PostProcessorsLists = append(build.PostProcessorsLists, []*PostProcessorBlock{pp})
		case buildPostProcessorsLabel:

			content, moreDiags := block.Body.Content(postProcessorsSchema)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}

			errored := false
			postProcessors := []*PostProcessorBlock{}
			for _, block := range content.Blocks {
				pp, moreDiags := p.decodePostProcessor(block, ectx)
				diags = append(diags, moreDiags...)
				if moreDiags.HasErrors() {
					errored = true
					break
				}
				postProcessors = append(postProcessors, pp)
			}
			if errored == false {
				build.PostProcessorsLists = append(build.PostProcessorsLists, postProcessors)
			}
		}
	}

	if !hadSource {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "missing source reference",
			Detail:   "a build block must reference at least one source to be built",
			Severity: hcl.DiagError,
			Subject:  block.DefRange.Ptr(),
		})
	}

	return build, diags
}
