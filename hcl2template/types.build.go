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
	// Sources is the list of sources that we want to start in this build block.
	Sources []SourceRef

	// ProvisionerBlocks references a list of HCL provisioner block that will
	// will be ran against the sources.
	ProvisionerBlocks []*ProvisionerBlock

	// ProvisionerBlocks references a list of HCL post-processors block that
	// will be ran against the artifacts from the provisioning steps.
	PostProcessors []*PostProcessorBlock

	HCL2Ref HCL2Ref
}

type Builds []*BuildBlock

// decodeBuildConfig is called when a 'build' block has been detected. It will
// load the references to the contents of the build block.
func (p *Parser) decodeBuildConfig(block *hcl.Block) (*BuildBlock, hcl.Diagnostics) {
	build := &BuildBlock{}

	var b struct {
		FromSources []string `hcl:"sources"`
		Config      hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)

	for _, buildFrom := range b.FromSources {
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

		build.Sources = append(build.Sources, ref)
	}

	content, moreDiags := b.Config.Content(buildSchema)
	diags = append(diags, moreDiags...)
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
			pp, moreDiags := p.decodePostProcessor(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			build.PostProcessors = append(build.PostProcessors, pp)
		}
	}

	return build, diags
}
