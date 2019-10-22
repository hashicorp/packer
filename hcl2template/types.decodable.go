package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type Decodable interface {
	FlatMapstructure() interface{}
}

type SelfSpecified interface {
	HCL2Spec() map[string]hcldec.Spec
}

func decodeDecodable(block *hcl.Block, ctx *hcl.EvalContext, dec Decodable) (interface{}, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	flatCfg := dec.FlatMapstructure()
	var spec hcldec.ObjectSpec
	if ss, selfSpecified := flatCfg.(SelfSpecified); selfSpecified {
		spec = hcldec.ObjectSpec(ss.HCL2Spec())
	} else {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Unknown type",
			Subject: &block.DefRange,
			Detail:  fmt.Sprintf("Cannot get spec from a %T", flatCfg),
		})
		return nil, diags
	}
	val, moreDiags := hcldec.Decode(block.Body, spec, ctx)
	diags = append(diags, moreDiags...)

	err := gocty.FromCtyValue(val, flatCfg)
	if err != nil {
		switch err := err.(type) {
		case cty.PathError:
			diags = append(diags, &hcl.Diagnostic{
				Summary: "gocty.FromCtyValue: " + err.Error(),
				Subject: &block.DefRange,
				Detail:  fmt.Sprintf("%v", err.Path),
			})
		default:
			diags = append(diags, &hcl.Diagnostic{
				Summary: "gocty.FromCtyValue: " + err.Error(),
				Subject: &block.DefRange,
				Detail:  fmt.Sprintf("%v", err),
			})
		}
	}
	return flatCfg, diags
}
