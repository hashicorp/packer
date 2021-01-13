package hcl2helper

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/mitchellh/mapstructure"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// UnknownVariableValue is a sentinel value that can be used
// to denote that the value of a variable is unknown at this time.
// RawConfig uses this information to build up data about
// unknown keys.
const UnknownVariableValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

// HCL2ValueFromConfigValue takes a value turns it into
// a cty.Value so it can be used within, for example, an HCL2 EvalContext.
func HCL2ValueFromConfigValue(v interface{}) cty.Value {
	if v == nil {
		return cty.NullVal(cty.DynamicPseudoType)
	}
	if v == UnknownVariableValue {
		return cty.DynamicVal
	}

	switch tv := v.(type) {
	case bool:
		return cty.BoolVal(tv)
	case string:
		return cty.StringVal(tv)
	case int:
		return cty.NumberIntVal(int64(tv))
	case float64:
		return cty.NumberFloatVal(tv)
	case []interface{}:
		vals := make([]cty.Value, len(tv))
		for i, ev := range tv {
			vals[i] = HCL2ValueFromConfigValue(ev)
		}
		return cty.TupleVal(vals)
	case []string:
		vals := make([]cty.Value, len(tv))
		for i, ev := range tv {
			vals[i] = cty.StringVal(ev)
		}
		return cty.ListVal(vals)
	case map[string]interface{}:
		vals := map[string]cty.Value{}
		for k, ev := range tv {
			vals[k] = HCL2ValueFromConfigValue(ev)
		}
		return cty.ObjectVal(vals)
	default:
		// HCL/HIL should never generate anything that isn't caught by
		// the above, so if we get here something has gone very wrong.
		panic(fmt.Errorf("can't convert %#v to cty.Value", v))
	}
}

// HCL2ValueFromConfig takes a struct with it's map of hcldec.Spec, and turns it into
// a cty.Value so it can be used as, for example, a Datasource value.
func HCL2ValueFromConfig(conf interface{}, configSpec map[string]hcldec.Spec) cty.Value {
	c := map[string]interface{}{}
	if err := mapstructure.Decode(conf, &c); err != nil {
		panic(fmt.Errorf("can't convert %#v to cty.Value", conf))
	}

	// Use the HCL2Spec to know the expected cty.Type for an attribute
	resp := map[string]cty.Value{}
	for k, v := range c {
		spec := configSpec[k]

		switch st := spec.(type) {
		case *hcldec.BlockListSpec:
			// This should be a slice of objects, so we need to take a special care
			if hcldec.ImpliedType(st.Nested).IsObjectType() {
				res := []cty.Value{}
				c := []interface{}{}
				if err := mapstructure.Decode(v, &c); err != nil {
					panic(fmt.Errorf("can't convert %#v to cty.Value", conf))
				}
				types := hcldec.ChildBlockTypes(spec)
				for _, e := range c {
					res = append(res, HCL2ValueFromConfig(e, types[k].(hcldec.ObjectSpec)))
				}
				if len(res) != 0 {
					resp[k] = cty.ListVal(res)
					continue
				}
				// At this point this is an empty list so we want it to go to gocty.ToCtyValue(v, impT)
				// and make it a NullVal
			}
		}

		impT := hcldec.ImpliedType(spec)
		if value, err := gocty.ToCtyValue(v, impT); err == nil {
			resp[k] = value
			continue
		}

		// Uncommon types not caught until now
		switch tv := v.(type) {
		case config.Trilean:
			resp[k] = cty.BoolVal(tv.True())
			continue
		case time.Duration:
			if tv.Microseconds() == int64(0) {
				resp[k] = cty.NumberIntVal(int64(0))
				continue
			}
			resp[k] = cty.NumberIntVal(v.(time.Duration).Milliseconds())
			continue
		}

		// This is a nested object and we should recursively go through the same process
		if impT.IsObjectType() {
			types := hcldec.ChildBlockTypes(spec)
			resp[k] = HCL2ValueFromConfig(v, types[k].(hcldec.ObjectSpec))
			continue
		}

		panic("not supported type - contact the Packer team with further information")
	}

	// This is decoding structs so it will always be an cty.ObjectVal at the end
	return cty.ObjectVal(resp)
}
