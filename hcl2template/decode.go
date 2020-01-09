package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

type Decodable interface {
	ConfigSpec() hcldec.ObjectSpec
}

func decodeHCL2Spec(block *hcl.Block, ectx *hcl.EvalContext, dec Decodable) (cty.Value, hcl.Diagnostics) {
	return hcldec.Decode(block.Body, dec.ConfigSpec(), ectx)
}
