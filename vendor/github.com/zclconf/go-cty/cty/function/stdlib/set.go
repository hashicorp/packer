package stdlib

import (
	"fmt"

	"github.com/zclconf/go-cty/cty/convert"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var SetHasElementFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "set",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
		{
			Name:             "elem",
			Type:             cty.DynamicPseudoType,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		return args[0].HasElement(args[1]), nil
	},
})

var SetUnionFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "first_set",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
	},
	VarParam: &function.Parameter{
		Name:             "other_sets",
		Type:             cty.Set(cty.DynamicPseudoType),
		AllowDynamicType: true,
	},
	Type: setOperationReturnType,
	Impl: setOperationImpl(func(s1, s2 cty.ValueSet) cty.ValueSet {
		return s1.Union(s2)
	}),
})

var SetIntersectionFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "first_set",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
	},
	VarParam: &function.Parameter{
		Name:             "other_sets",
		Type:             cty.Set(cty.DynamicPseudoType),
		AllowDynamicType: true,
	},
	Type: setOperationReturnType,
	Impl: setOperationImpl(func(s1, s2 cty.ValueSet) cty.ValueSet {
		return s1.Intersection(s2)
	}),
})

var SetSubtractFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "a",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
		{
			Name:             "b",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
	},
	Type: setOperationReturnType,
	Impl: setOperationImpl(func(s1, s2 cty.ValueSet) cty.ValueSet {
		return s1.Subtract(s2)
	}),
})

var SetSymmetricDifferenceFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "first_set",
			Type:             cty.Set(cty.DynamicPseudoType),
			AllowDynamicType: true,
		},
	},
	VarParam: &function.Parameter{
		Name:             "other_sets",
		Type:             cty.Set(cty.DynamicPseudoType),
		AllowDynamicType: true,
	},
	Type: setOperationReturnType,
	Impl: setOperationImpl(func(s1, s2 cty.ValueSet) cty.ValueSet {
		return s1.Subtract(s2)
	}),
})

// SetHasElement determines whether the given set contains the given value as an
// element.
func SetHasElement(set cty.Value, elem cty.Value) (cty.Value, error) {
	return SetHasElementFunc.Call([]cty.Value{set, elem})
}

// SetUnion returns a new set containing all of the elements from the given
// sets, which must have element types that can all be converted to some
// common type using the standard type unification rules. If conversion
// is not possible, an error is returned.
//
// The union operation is performed after type conversion, which may result
// in some previously-distinct values being conflated.
//
// At least one set must be provided.
func SetUnion(sets ...cty.Value) (cty.Value, error) {
	return SetUnionFunc.Call(sets)
}

// Intersection returns a new set containing the elements that exist
// in all of the given sets, which must have element types that can all be
// converted to some common type using the standard type unification rules.
// If conversion is not possible, an error is returned.
//
// The intersection operation is performed after type conversion, which may
// result in some previously-distinct values being conflated.
//
// At least one set must be provided.
func SetIntersection(sets ...cty.Value) (cty.Value, error) {
	return SetIntersectionFunc.Call(sets)
}

// SetSubtract returns a new set containing the elements from the
// first set that are not present in the second set. The sets must have
// element types that can both be converted to some common type using the
// standard type unification rules. If conversion is not possible, an error
// is returned.
//
// The subtract operation is performed after type conversion, which may
// result in some previously-distinct values being conflated.
func SetSubtract(a, b cty.Value) (cty.Value, error) {
	return SetSubtractFunc.Call([]cty.Value{a, b})
}

// SetSymmetricDifference returns a new set containing elements that appear
// in any of the given sets but not multiple. The sets must have
// element types that can all be converted to some common type using the
// standard type unification rules. If conversion is not possible, an error
// is returned.
//
// The difference operation is performed after type conversion, which may
// result in some previously-distinct values being conflated.
func SetSymmetricDifference(sets ...cty.Value) (cty.Value, error) {
	return SetSymmetricDifferenceFunc.Call(sets)
}

func setOperationReturnType(args []cty.Value) (ret cty.Type, err error) {
	var etys []cty.Type
	for _, arg := range args {
		etys = append(etys, arg.Type().ElementType())
	}
	newEty, _ := convert.UnifyUnsafe(etys)
	if newEty == cty.NilType {
		return cty.NilType, fmt.Errorf("given sets must all have compatible element types")
	}
	return cty.Set(newEty), nil
}

func setOperationImpl(f func(s1, s2 cty.ValueSet) cty.ValueSet) function.ImplFunc {
	return func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		first := args[0]
		first, err = convert.Convert(first, retType)
		if err != nil {
			return cty.NilVal, function.NewArgError(0, err)
		}

		set := first.AsValueSet()
		for i, arg := range args[1:] {
			arg, err := convert.Convert(arg, retType)
			if err != nil {
				return cty.NilVal, function.NewArgError(i+1, err)
			}

			argSet := arg.AsValueSet()
			set = f(set, argSet)
		}
		return cty.SetValFromValueSet(set), nil
	}
}
