package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const (
	buildFromLabel = "from"

	buildSourceLabel = "source"

	buildProvisionerLabel = "provisioner"

	buildPostProcessorLabel = "post-processor"
)

var buildSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: buildFromLabel, LabelNames: []string{"type"}},
		{Type: buildProvisionerLabel, LabelNames: []string{"type"}},
		{Type: buildPostProcessorLabel, LabelNames: []string{"type"}},
	},
}

type BuildBlock struct {
	ProvisionerBlocks []*ProvisionerBlock

	PostProcessors []*PostProcessorBlock

	Froms []SourceRef

	HCL2Ref HCL2Ref
}

type Builds []*BuildBlock

func (p *Parser) decodeBuildConfig(block *hcl.Block) (*BuildBlock, hcl.Diagnostics) {
	build := &BuildBlock{}

	var b struct {
		FromSources []string `hcl:"sources"`
		Config      hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)

	for _, buildFrom := range b.FromSources {
		ref := sourceRefFromString(buildFrom)

		if ref == NoSource {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid " + sourceLabel + " reference",
				Detail: "A " + sourceLabel + " type must start with a letter and " +
					"may contain only letters, digits, underscores, and dashes." +
					"A valid source reference looks like: `source.type.name`",
				Subject: &block.LabelRanges[0],
			})
			continue
		}
		if !hclsyntax.ValidIdentifier(ref.Type) ||
			!hclsyntax.ValidIdentifier(ref.Name) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid " + sourceLabel + " reference",
				Detail: "A " + sourceLabel + " type must start with a letter and " +
					"may contain only letters, digits, underscores, and dashes." +
					"A valid source reference looks like: `source.type.name`",
				Subject: &block.LabelRanges[0],
			})
			continue
		}

		build.Froms = append(build.Froms, ref)
	}

	content, diags := b.Config.Content(buildSchema)
	for _, block := range content.Blocks {
		switch block.Type {
		case buildProvisionerLabel:
			p, moreDiags := p.decodeProvisioner(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.ProvisionerBlocks = append(build.ProvisionerBlocks, p)
		case buildPostProcessorLabel:
			pp, moreDiags := p.decodePostProcessorGroup(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.PostProcessors = append(build.PostProcessors, pp)
		}
	}

	return build, diags
}
