// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// AnyTrue constructs a function that returns true if a single element of
// the list is true. If the list is empty, return false.
var AnyTrue = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "list",
			Type: cty.List(cty.Bool),
		},
	},
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		result := cty.False
		var hasUnknown bool
		for it := args[0].ElementIterator(); it.Next(); {
			_, v := it.Element()
			if !v.IsKnown() {
				hasUnknown = true
				continue
			}
			if v.IsNull() {
				continue
			}
			result = result.Or(v)
			if result.True() {
				return cty.True, nil
			}
		}
		if hasUnknown {
			return cty.UnknownVal(cty.Bool), nil
		}
		return result, nil
	},
})
