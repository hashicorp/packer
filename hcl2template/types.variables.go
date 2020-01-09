package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type PackerV1Variables map[string]cty.Value

// decodeConfig decodes a "variables" section the way packer 1 used to
func (variables *PackerV1Variables) decodeConfig(block *hcl.Block) hcl.Diagnostics {
	attrs, diags := block.Body.JustAttributes()

	if diags.HasErrors() {
		return diags
	}

	for key, attr := range attrs {
		if _, found := (*variables)[key]; found {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate variable",
				Detail:   "Duplicate " + key + " variable found.",
				Subject:  attr.NameRange.Ptr(),
				Context:  block.DefRange.Ptr(),
			})
			continue
		}
		value, moreDiags := attr.Expr.Value(nil)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		(*variables)[key] = value
	}

	return diags
}
