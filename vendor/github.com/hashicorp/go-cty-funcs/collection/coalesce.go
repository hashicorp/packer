package collection

import (
	"errors"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

// Coalesce takes any number of arguments and returns the first one that isn't
// empty.
func Coalesce(args ...cty.Value) (cty.Value, error) {
	return CoalesceFunc.Call(args)
}

// CoalesceFunc constructs a function that takes any number of arguments and
// returns the first one that isn't empty. This function was copied from go-cty
// stdlib and modified so that it returns the first *non-empty* non-null
// element from a sequence, instead of merely the first non-null.
var CoalesceFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name:             "vals",
		Type:             cty.DynamicPseudoType,
		AllowUnknown:     true,
		AllowDynamicType: true,
		AllowNull:        true,
	},
	Type: func(args []cty.Value) (ret cty.Type, err error) {
		argTypes := make([]cty.Type, len(args))
		for i, val := range args {
			argTypes[i] = val.Type()
		}
		retType, _ := convert.UnifyUnsafe(argTypes)
		if retType == cty.NilType {
			return cty.NilType, errors.New("all arguments must have the same type")
		}
		return retType, nil
	},
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		for _, argVal := range args {
			// We already know this will succeed because of the checks in our
			// Type func above
			argVal, _ = convert.Convert(argVal, retType)
			if !argVal.IsKnown() {
				return cty.UnknownVal(retType), nil
			}
			if argVal.IsNull() {
				continue
			}
			if retType == cty.String && argVal.RawEquals(cty.StringVal("")) {
				continue
			}

			return argVal, nil
		}
		return cty.NilVal, errors.New("no non-null, non-empty-string arguments")
	},
})
