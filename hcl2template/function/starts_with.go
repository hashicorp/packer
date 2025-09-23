package function

import (
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// StartsWithFunc constructs a function that checks if a string starts with
// a specific prefix using strings.HasPrefix
var StartsWithFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:         "str",
			Type:         cty.String,
			AllowUnknown: false,
		},
		{
			Name: "prefix",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		str := args[0].AsString()
		prefix := args[1].AsString()

		return cty.BoolVal(strings.HasPrefix(str, prefix)), nil
	},
})
