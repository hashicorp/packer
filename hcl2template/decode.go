package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type Decodable interface {
	ConfigSpec() hcldec.ObjectSpec
}

func decodeHCL2Spec(body hcl.Body, ctx *hcl.EvalContext, dec Decodable) (cty.Value, hcl.Diagnostics) {
	return hcldec.Decode(body, dec.ConfigSpec(), ctx)
}
