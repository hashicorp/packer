package hcl2template

import (
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type BuildFromList []BuildFrom

type BuildFrom struct {
	// source to take config from
	Src SourceRef `hcl:"-"`

	HCL2Ref HCL2Ref
}

func sourceRefFromString(in string) SourceRef {
	args := strings.Split(in, ".")
	if len(args) < 2 {
		return NoSource
	}
	if len(args) > 2 {
		// src.type.name
		args = args[1:]
	}
	return SourceRef{
		Type: args[0],
		Name: args[1],
	}
}

func (bf *BuildFrom) decodeConfig(block *hcl.Block) hcl.Diagnostics {

	bf.Src = sourceRefFromString(block.Labels[0])
	bf.HCL2Ref.DeclRange = block.DefRange

	var b struct {
		Config hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)

	if bf.Src == NoSource {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid " + sourceLabel + " reference",
			Detail: "A " + sourceLabel + " type must start with a letter and " +
				"may contain only letters, digits, underscores, and dashes." +
				"A valid source reference looks like: `src.type.name`",
			Subject: &block.LabelRanges[0],
		})
	}
	if !hclsyntax.ValidIdentifier(bf.Src.Type) ||
		!hclsyntax.ValidIdentifier(bf.Src.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid " + sourceLabel + " reference",
			Detail: "A " + sourceLabel + " type must start with a letter and " +
				"may contain only letters, digits, underscores, and dashes." +
				"A valid source reference looks like: `src.type.name`",
			Subject: &block.LabelRanges[0],
		})
	}

	return diags
}
