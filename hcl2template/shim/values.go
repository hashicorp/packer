package hcl2shim

import (
	"fmt"
	"math/big"

	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/zclconf/go-cty/cty"
)

// ConfigValueFromHCL2 converts a value from HCL2 (really, from the cty dynamic
// types library that HCL2 uses) to a value type that matches what would've
// been produced from the HCL-based interpolator for an equivalent structure.
//
// This function will transform a cty null value into a Go nil value, which
// isn't a possible outcome of the HCL/HIL-based decoder and so callers may
// need to detect and reject any null values.
func ConfigValueFromHCL2(v cty.Value) interface{} {
	if !v.IsKnown() {
		return hcl2helper.UnknownVariableValue
	}
	if v.IsNull() {
		return nil
	}

	switch v.Type() {
	case cty.Bool:
		return v.True() // like HCL.BOOL
	case cty.String:
		return v.AsString() // like HCL token.STRING or token.HEREDOC
	case cty.Number:
		// We can't match HCL _exactly_ here because it distinguishes between
		// int and float values, but we'll get as close as we can by using
		// an int if the number is exactly representable, and a float if not.
		// The conversion to float will force precision to that of a float64,
		// which is potentially losing information from the specific number
		// given, but no worse than what HCL would've done in its own conversion
		// to float.

		f := v.AsBigFloat()
		if i, acc := f.Int64(); acc == big.Exact {
			// if we're on a 32-bit system and the number is too big for 32-bit
			// int then we'll fall through here and use a float64.
			const MaxInt = int(^uint(0) >> 1)
			const MinInt = -MaxInt - 1
			if i <= int64(MaxInt) && i >= int64(MinInt) {
				return int(i) // Like HCL token.NUMBER
			}
		}

		f64, _ := f.Float64()
		return f64 // like HCL token.FLOAT
	}

	if v.Type().IsListType() || v.Type().IsSetType() || v.Type().IsTupleType() {
		l := make([]interface{}, 0, v.LengthInt())
		it := v.ElementIterator()
		for it.Next() {
			_, ev := it.Element()
			l = append(l, ConfigValueFromHCL2(ev))
		}
		return l
	}

	if v.Type().IsMapType() || v.Type().IsObjectType() {
		l := make(map[string]interface{})
		it := v.ElementIterator()
		for it.Next() {
			ek, ev := it.Element()
			cv := ConfigValueFromHCL2(ev)
			if cv != nil {
				l[ek.AsString()] = cv
			}
		}
		return l
	}

	// If we fall out here then we have some weird type that we haven't
	// accounted for. This should never happen unless the caller is using
	// capsule types, and we don't currently have any such types defined.
	panic(fmt.Errorf("can't convert %#v to config value", v))
}

// WriteUnknownPlaceholderValues will replace every Unknown value with a equivalent placeholder.
// This is useful to use before marshaling the value to JSON. The default values are:
// - string: "<unknown>"
// - number: 0
// - bool: false
// - objects/lists/tuples/sets/maps: empty
func WriteUnknownPlaceholderValues(v cty.Value) cty.Value {
	if v.IsNull() {
		return v
	}
	t := v.Type()
	switch {
	case t.IsPrimitiveType():
		if v.IsKnown() {
			return v
		}
		switch t {
		case cty.String:
			return cty.StringVal("<unknown>")
		case cty.Number:
			return cty.MustParseNumberVal("0")
		case cty.Bool:
			return cty.BoolVal(false)
		default:
			panic("unsupported primitive type")
		}
	case t.IsListType():
		if !v.IsKnown() {
			return cty.ListValEmpty(t.ElementType())
		}
		arr := []cty.Value{}
		it := v.ElementIterator()
		for it.Next() {
			_, ev := it.Element()
			arr = append(arr, WriteUnknownPlaceholderValues(ev))
		}
		if len(arr) == 0 {
			return cty.ListValEmpty(t.ElementType())
		}
		return cty.ListVal(arr)
	case t.IsSetType():
		if !v.IsKnown() {
			return cty.SetValEmpty(t.ElementType())
		}
		arr := []cty.Value{}
		it := v.ElementIterator()
		for it.Next() {
			_, ev := it.Element()
			arr = append(arr, WriteUnknownPlaceholderValues(ev))
		}
		if len(arr) == 0 {
			return cty.SetValEmpty(t.ElementType())
		}
		return cty.SetVal(arr)
	case t.IsMapType():
		if !v.IsKnown() {
			return cty.MapValEmpty(t.ElementType())
		}
		obj := map[string]cty.Value{}
		it := v.ElementIterator()
		for it.Next() {
			ek, ev := it.Element()
			obj[ek.AsString()] = WriteUnknownPlaceholderValues(ev)
		}
		if len(obj) == 0 {
			return cty.MapValEmpty(t.ElementType())
		}
		return cty.MapVal(obj)
	case t.IsTupleType():
		if !v.IsKnown() {
			return cty.EmptyTupleVal
		}
		arr := []cty.Value{}
		it := v.ElementIterator()
		for it.Next() {
			_, ev := it.Element()
			arr = append(arr, WriteUnknownPlaceholderValues(ev))
		}
		if len(arr) == 0 {
			return cty.EmptyTupleVal
		}
		return cty.TupleVal(arr)
	case t.IsObjectType():
		if !v.IsKnown() {
			return cty.EmptyObjectVal
		}
		obj := map[string]cty.Value{}
		it := v.ElementIterator()
		for it.Next() {
			ek, ev := it.Element()
			obj[ek.AsString()] = WriteUnknownPlaceholderValues(ev)
		}
		if len(obj) == 0 {
			return cty.EmptyObjectVal
		}
		return cty.ObjectVal(obj)
	default:
		// should never happen
		panic("unknown type")
	}
}
