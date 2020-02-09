package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// Decodable structs are structs that can tell their hcl2 ObjectSpec; this
// config spec will be passed to hcldec.Decode and the result will be a
// cty.Value. This Value can then be applied on the said struct.
type Decodable interface {
	ConfigSpec() hcldec.ObjectSpec
}

func decodeHCL2Spec(body hcl.Body, ectx *hcl.EvalContext, dec Decodable) (cty.Value, hcl.Diagnostics) {
	return hcldec.Decode(body, dec.ConfigSpec(), ectx)
}
