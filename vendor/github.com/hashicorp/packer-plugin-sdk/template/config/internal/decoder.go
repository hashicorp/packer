package internal

import (
	"fmt"
	"reflect"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// HCL is not able to decode map[string]interface{}, we must use a map[string]cty.Value instead.
// ref: https://github.com/hashicorp/hcl/issues/291#issuecomment-496347585
// MapOfInterfaceToMapOfCTY will transform the map[string]interface{} decoded from a config
// to the expected map[string]cty.Value.
func MapOfInterfaceToMapOfCTY(f reflect.Type, t reflect.Type, v interface{}) (interface{}, error) {
	if t == reflect.TypeOf(map[string]cty.Value{}) {
		to := map[string]cty.Value{}
		if from, ok := v.(map[string]interface{}); ok {
			for key, val := range from {
				to[key] = InterfaceToCTY(val)
			}
		}
		return to, nil
	}
	return v, nil
}

// InterfaceToCTY is a similar to hcl2helper.HCL2ValueFromConfigValue
// and is placed here to avoid cyclic dependency.
// It takes a value and turns it into a cty.Value for the config decoder.
// This for internal usage only, used by the decoder above.
func InterfaceToCTY(v interface{}) cty.Value {
	if v == nil {
		return cty.NullVal(cty.DynamicPseudoType)
	}

	switch tv := v.(type) {
	case []interface{}:
		vals := make([]cty.Value, len(tv))
		for i, ev := range tv {
			vals[i] = InterfaceToCTY(ev)
		}
		return cty.TupleVal(vals)
	case []string:
		vals := make([]cty.Value, len(tv))
		for i, ev := range tv {
			vals[i] = cty.StringVal(ev)
		}
		if len(vals) == 0 {
			return cty.ListValEmpty(cty.String)
		}
		return cty.ListVal(vals)
	case map[string]interface{}:
		vals := map[string]cty.Value{}
		for k, ev := range tv {
			vals[k] = InterfaceToCTY(ev)
		}
		if len(vals) == 0 {
			return cty.MapValEmpty(cty.String)
		}
		return cty.MapVal(vals)
	}

	impliedValType, err := gocty.ImpliedType(v)
	if err != nil {
		// HCL/HIL should never generate anything that isn't caught by
		// the above, so if we get here something has gone very wrong.
		panic(fmt.Errorf("can't convert %#v to cty.Value: %s", v, err.Error()))
	}
	value, err := gocty.ToCtyValue(v, impliedValType)
	if err != nil {
		// HCL/HIL should never generate anything that isn't caught by
		// the above, so if we get here something has gone very wrong.
		panic(fmt.Errorf("can't convert %#v to cty.Value %s", v, err.Error()))
	}
	return value
}
