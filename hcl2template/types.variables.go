package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/zclconf/go-cty/cty"
)

type Variable struct {
	Default     cty.Value
	Type        cty.Type
	Description string
	Sensible    bool

	block *hcl.Block
}

func (v *Variable) Value() cty.Value {
	return v.Default
}

type Variables map[string]Variable

func (variables Variables) Values() map[string]cty.Value {
	res := map[string]cty.Value{}
	for k, v := range variables {
		res[k] = v.Value()
	}
	return res
}

// decodeConfig decodes a "variables" section the way packer 1 used to
func (variables *Variables) decodeConfigMap(block *hcl.Block, ectx *hcl.EvalContext) hcl.Diagnostics {
	if (*variables) == nil {
		(*variables) = Variables{}
	}
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
		value, moreDiags := attr.Expr.Value(ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}
		(*variables)[key] = Variable{
			Default: value,
			Type:    value.Type(),
		}
	}

	return diags
}

// decodeConfig decodes a "variables" section the way packer 1 used to
func (variables *Variables) decodeConfig(block *hcl.Block, ectx *hcl.EvalContext) hcl.Diagnostics {
	if (*variables) == nil {
		(*variables) = Variables{}
	}

	var b struct {
		Description string   `hcl:"description,optional"`
		Sensible    bool     `hcl:"sensible,optional"`
		Rest        hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)

	if diags.HasErrors() {
		return diags
	}

	res := Variable{
		Description: b.Description,
		Sensible:    b.Sensible,
		block:       block,
	}

	attrs, moreDiags := b.Rest.JustAttributes()
	diags = append(diags, moreDiags...)

	if def, ok := attrs["default"]; ok {
		defaultValue, moreDiags := def.Expr.Value(ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			return diags
		}
		res.Default = defaultValue
		res.Type = defaultValue.Type()
	}
	if t, ok := attrs["type"]; ok {
		tp, moreDiags := typeexpr.Type(t.Expr)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			return diags
		}

		res.Type = tp
	}

	(*variables)[block.Labels[0]] = res

	return diags
}
