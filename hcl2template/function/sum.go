package function

import (
	"fmt"
	"math/big"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
)

var SumFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "list",
			Type: cty.DynamicPseudoType,
		},
	},
	Type:         function.StaticReturnType(cty.Number),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {

		if !args[0].CanIterateElements() {
			return cty.NilVal, function.NewArgErrorf(0, "cannot sum noniterable")
		}

		if args[0].LengthInt() == 0 { // Easy path
			return cty.NilVal, function.NewArgErrorf(0, "cannot sum an empty list")
		}

		arg := args[0].AsValueSlice()
		ty := args[0].Type()

		if !ty.IsListType() && !ty.IsSetType() && !ty.IsTupleType() {
			return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple. Received %s", ty.FriendlyName())
		}

		if !args[0].IsWhollyKnown() {
			return cty.UnknownVal(cty.Number), nil
		}

		// big.Float.Add can panic if the input values are opposing infinities,
		// so we must catch that here in order to remain within
		// the cty Function abstraction.
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(big.ErrNaN); ok {
					ret = cty.NilVal
					err = fmt.Errorf("can't compute sum of opposing infinities")
				} else {
					// not a panic we recognize
					panic(r)
				}
			}
		}()

		s := arg[0]
		if s.IsNull() {
			return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
		}
		s, err = convert.Convert(s, cty.Number)
		if err != nil {
			return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
		}
		for _, v := range arg[1:] {
			if v.IsNull() {
				return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
			}
			v, err = convert.Convert(v, cty.Number)
			if err != nil {
				return cty.NilVal, function.NewArgErrorf(0, "argument must be list, set, or tuple of number values")
			}
			s = s.Add(v)
		}

		return s, nil
	},
})

// Sum adds numbers in a list, set, or tuple
func Sum(list cty.Value) (cty.Value, error) {
	return SumFunc.Call([]cty.Value{list})
}
