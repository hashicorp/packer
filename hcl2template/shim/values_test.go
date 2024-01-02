// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2shim

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	"github.com/zclconf/go-cty/cty"
)

func TestConfigValueFromHCL2(t *testing.T) {
	tests := []struct {
		Input cty.Value
		Want  interface{}
	}{
		{
			cty.True,
			true,
		},
		{
			cty.False,
			false,
		},
		{
			cty.NumberIntVal(12),
			int(12),
		},
		{
			cty.NumberFloatVal(12.5),
			float64(12.5),
		},
		{
			cty.StringVal("hello world"),
			"hello world",
		},
		{
			cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Ermintrude"),
				"age":  cty.NumberIntVal(19),
				"address": cty.ObjectVal(map[string]cty.Value{
					"street": cty.ListVal([]cty.Value{cty.StringVal("421 Shoreham Loop")}),
					"city":   cty.StringVal("Fridgewater"),
					"state":  cty.StringVal("MA"),
					"zip":    cty.StringVal("91037"),
				}),
			}),
			map[string]interface{}{
				"name": "Ermintrude",
				"age":  int(19),
				"address": map[string]interface{}{
					"street": []interface{}{"421 Shoreham Loop"},
					"city":   "Fridgewater",
					"state":  "MA",
					"zip":    "91037",
				},
			},
		},
		{
			cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"bar": cty.StringVal("baz"),
			}),
			map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
		},
		{
			cty.TupleVal([]cty.Value{
				cty.StringVal("foo"),
				cty.True,
			}),
			[]interface{}{
				"foo",
				true,
			},
		},
		{
			cty.NullVal(cty.String),
			nil,
		},
		{
			cty.UnknownVal(cty.String),
			hcl2helper.UnknownVariableValue,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v", test.Input), func(t *testing.T) {
			got := ConfigValueFromHCL2(test.Input)
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v", test.Input, got, test.Want)
			}
		})
	}
}

func TestWriteUnknownPlaceholderValues(t *testing.T) {
	tests := []struct {
		Name  string
		Input cty.Value
		Want  cty.Value
	}{
		{
			Name:  "Unknown bool",
			Input: cty.UnknownVal(cty.Bool),
			Want:  cty.False,
		},
		{
			Name:  "Unknown number",
			Input: cty.UnknownVal(cty.Number),
			Want:  cty.NumberIntVal(0),
		},
		{
			Name:  "Unknown string",
			Input: cty.UnknownVal(cty.String),
			Want:  cty.StringVal("<unknown>"),
		},
		{
			Name:  "Unknown object",
			Input: cty.UnknownVal(cty.EmptyObject),
			Want:  cty.EmptyObjectVal,
		},
		{
			Name: "Object with unknown values",
			Input: cty.ObjectVal(map[string]cty.Value{
				"name":    cty.UnknownVal(cty.String),
				"address": cty.UnknownVal(cty.EmptyObject),
			}),
			Want: cty.ObjectVal(map[string]cty.Value{
				"name":    cty.StringVal("<unknown>"),
				"address": cty.EmptyObjectVal,
			}),
		},
		{
			Name:  "Empty object",
			Input: cty.ObjectVal(map[string]cty.Value{}),
			Want:  cty.EmptyObjectVal,
		},
		{
			Name:  "Unknown tuple",
			Input: cty.UnknownVal(cty.EmptyTuple),
			Want:  cty.EmptyTupleVal,
		},
		{
			Name: "Tuple with unknown values",
			Input: cty.TupleVal([]cty.Value{
				cty.UnknownVal(cty.String),
				cty.UnknownVal(cty.Bool),
			}),
			Want: cty.TupleVal([]cty.Value{
				cty.StringVal("<unknown>"),
				cty.False,
			}),
		},
		{
			Name:  "Empty tuple",
			Input: cty.TupleVal([]cty.Value{}),
			Want:  cty.EmptyTupleVal,
		},
		{
			Name:  "Unknown list",
			Input: cty.UnknownVal(cty.List(cty.String)),
			Want:  cty.ListValEmpty(cty.String),
		},
		{
			Name: "List with unknown values",
			Input: cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("<unknown>"),
			}),
		},
		{
			Name:  "Empty list",
			Input: cty.ListValEmpty(cty.String),
			Want:  cty.ListValEmpty(cty.String),
		},
		{
			Name:  "Unknown set",
			Input: cty.UnknownVal(cty.Set(cty.String)),
			Want:  cty.SetValEmpty(cty.String),
		},
		{
			Name: "Set with unknown values",
			Input: cty.SetVal([]cty.Value{
				cty.UnknownVal(cty.String),
			}),
			Want: cty.SetVal([]cty.Value{
				cty.StringVal("<unknown>"),
			}),
		},
		{
			Name:  "Empty Set",
			Input: cty.SetValEmpty(cty.String),
			Want:  cty.SetValEmpty(cty.String),
		},
		{
			Name:  "Unknown map",
			Input: cty.UnknownVal(cty.Map(cty.String)),
			Want:  cty.MapValEmpty(cty.String),
		},
		{
			Name: "Map with unknown values",
			Input: cty.MapVal(map[string]cty.Value{
				"name": cty.UnknownVal(cty.String),
			}),
			Want: cty.MapVal(map[string]cty.Value{
				"name": cty.StringVal("<unknown>"),
			}),
		},
		{
			Name:  "Empty Map",
			Input: cty.MapValEmpty(cty.String),
			Want:  cty.MapValEmpty(cty.String),
		},
		{
			Name:  "Null val",
			Input: cty.NullVal(cty.String),
			Want:  cty.NullVal(cty.String),
		},
	}

	for _, test := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			got := WriteUnknownPlaceholderValues(test.Input)
			if got.Equals(test.Want).False() {
				t.Errorf("wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v", test.Input, got, test.Want)
			}
		})
	}
}
