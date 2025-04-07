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
			AllowUnknown: true,
		},
		{
			Name: "prefix",
			Type: cty.String,
		},
	},
	Type:         function.StaticReturnType(cty.Bool),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		prefix := args[1].AsString()

		if !args[0].IsKnown() {
			// If the unknown value has a known prefix then we might be
			// able to still produce a known result.
			if prefix == "" {
				// The empty string is a prefix of any string.
				return cty.True, nil
			}
			if knownPrefix := args[0].Range().StringPrefix(); knownPrefix != "" {
				if strings.HasPrefix(knownPrefix, prefix) {
					return cty.True, nil
				}
				if len(knownPrefix) >= len(prefix) {
					// If the prefix we're testing is no longer than the known
					// prefix and it didn't match then the full string with
					// that same prefix can't match either.
					return cty.False, nil
				}
			}
			return cty.UnknownVal(cty.Bool), nil
		}

		str := args[0].AsString()

		if strings.HasPrefix(str, prefix) {
			return cty.True, nil
		}

		return cty.False, nil
	},
})
