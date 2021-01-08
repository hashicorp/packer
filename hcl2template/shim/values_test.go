package hcl2shim

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
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
			UnknownVariableValue,
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

func TestHCL2ValueFromConfigValue(t *testing.T) {
	tests := []struct {
		Name  string
		Input interface{}
		Want  cty.Value
	}{
		{
			Name:  "bool true",
			Input: true,
			Want:  cty.True,
		},
		{
			Name:  "bool false",
			Input: false,
			Want:  cty.False,
		},
		{
			Name:  "int",
			Input: int(12),
			Want:  cty.NumberIntVal(12),
		},
		{
			Name:  "float64",
			Input: float64(12.5),
			Want:  cty.NumberFloatVal(12.5),
		},
		{
			Name:  "string",
			Input: "hello world",
			Want:  cty.StringVal("hello world"),
		},
		{
			Name: "nested map[string]interface{}",
			Input: map[string]interface{}{
				"name": "Ermintrude",
				"age":  int(19),
				"address": map[string]interface{}{
					"street": []interface{}{"421 Shoreham Loop"},
					"city":   "Fridgewater",
					"state":  "MA",
					"zip":    "91037",
				},
			},
			Want: cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("Ermintrude"),
				"age":  cty.NumberIntVal(19),
				"address": cty.ObjectVal(map[string]cty.Value{
					"street": cty.TupleVal([]cty.Value{cty.StringVal("421 Shoreham Loop")}),
					"city":   cty.StringVal("Fridgewater"),
					"state":  cty.StringVal("MA"),
					"zip":    cty.StringVal("91037"),
				}),
			}),
		},
		{
			Name: "simple map[string]interface{}",
			Input: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Want: cty.ObjectVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
				"bar": cty.StringVal("baz"),
			}),
		},
		{
			Name: "[]interface{} as tuple",
			Input: []interface{}{
				"foo",
				true,
			},
			Want: cty.TupleVal([]cty.Value{
				cty.StringVal("foo"),
				cty.True,
			}),
		},
		{
			Name:  "nil",
			Input: nil,
			Want:  cty.NullVal(cty.DynamicPseudoType),
		},
		{
			Name:  "UnknownVariableValue",
			Input: UnknownVariableValue,
			Want:  cty.DynamicVal,
		},
		{
			Name: "SliceString",
			Input: []string{
				"a",
				"b",
				"c",
			},
			Want: cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := HCL2ValueFromConfigValue(test.Input)
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v", test.Input, got, test.Want)
			}
		})
	}
}

func TestHCL2ValueFromConfig(t *testing.T) {
	tests := []struct {
		Name  string
		Input interface{}
		Spec  map[string]hcldec.Spec
		Want  cty.Value
	}{
		{
			Name:  "Empty config",
			Input: MockConfig{},
			Spec:  new(MockConfig).FlatMapstructure().HCL2Spec(),
			Want: cty.ObjectVal(map[string]cty.Value{
				"not_squashed":            cty.StringVal(""),
				"string":                  cty.StringVal(""),
				"int":                     cty.NumberIntVal(int64(0)),
				"int64":                   cty.NumberIntVal(int64(0)),
				"bool":                    cty.False,
				"trilean":                 cty.False,
				"duration":                cty.NumberIntVal(int64(0)),
				"map_string_string":       cty.NullVal(cty.Map(cty.String)),
				"slice_string":            cty.NullVal(cty.List(cty.String)),
				"slice_slice_string":      cty.NullVal(cty.List(cty.List(cty.String))),
				"named_map_string_string": cty.NullVal(cty.Map(cty.String)),
				"named_string":            cty.StringVal(""),
				"tag": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
					"key": cty.String, "value": cty.String,
				}))),
				"data_source": cty.StringVal(""),
				"nested": cty.ObjectVal(map[string]cty.Value{
					"string":                  cty.StringVal(""),
					"int":                     cty.NumberIntVal(int64(0)),
					"int64":                   cty.NumberIntVal(int64(0)),
					"bool":                    cty.False,
					"trilean":                 cty.False,
					"duration":                cty.NumberIntVal(int64(0)),
					"map_string_string":       cty.NullVal(cty.Map(cty.String)),
					"slice_string":            cty.NullVal(cty.List(cty.String)),
					"slice_slice_string":      cty.NullVal(cty.List(cty.List(cty.String))),
					"named_map_string_string": cty.NullVal(cty.Map(cty.String)),
					"named_string":            cty.StringVal(""),
					"tag": cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
						"key": cty.String, "value": cty.String,
					}))),
					"data_source": cty.StringVal(""),
				}),
				"nested_slice": cty.NullVal(hcldec.ImpliedType(new(MockConfig).FlatMapstructure().HCL2Spec()["nested_slice"])),
			}),
		},
		{
			Name: "Full filled config",
			Input: MockConfig{
				NotSquashed: "not squashed",
				NestedMockConfig: NestedMockConfig{
					String:               "string",
					Int:                  1,
					Int64:                int64(2),
					Bool:                 true,
					Trilean:              config.TriTrue,
					Duration:             10 * time.Second,
					MapStringString:      map[string]string{"a": "b"},
					SliceString:          []string{"a", "b"},
					SliceSliceString:     [][]string{{"a", "b"}},
					NamedMapStringString: NamedMapStringString{"a": "b"},
					NamedString:          "named string",
					Tags: []MockTag{{
						Key:   "a",
						Value: "b",
					}},
					Datasource: "datasource",
				},
				Nested: NestedMockConfig{
					String:               "string",
					Int:                  1,
					Int64:                int64(2),
					Bool:                 true,
					Trilean:              config.TriTrue,
					Duration:             10 * time.Second,
					MapStringString:      map[string]string{"a": "b"},
					SliceString:          []string{"a", "b"},
					SliceSliceString:     [][]string{{"a", "b"}},
					NamedMapStringString: NamedMapStringString{"a": "b"},
					NamedString:          "named string",
					Tags: []MockTag{{
						Key:   "a",
						Value: "b",
					}},
					Datasource: "datasource",
				},
				NestedSlice: []NestedMockConfig{
					{
						String:               "string",
						Int:                  1,
						Int64:                int64(2),
						Bool:                 true,
						Trilean:              config.TriTrue,
						Duration:             10 * time.Second,
						MapStringString:      map[string]string{"a": "b"},
						SliceString:          []string{"a", "b"},
						SliceSliceString:     [][]string{{"a", "b"}},
						NamedMapStringString: NamedMapStringString{"a": "b"},
						NamedString:          "named string",
						Tags: []MockTag{{
							Key:   "a",
							Value: "b",
						}},
						Datasource: "datasource",
					},
				},
			},
			Spec: new(MockConfig).FlatMapstructure().HCL2Spec(),
			Want: cty.ObjectVal(map[string]cty.Value{
				"not_squashed":      cty.StringVal("not squashed"),
				"string":            cty.StringVal("string"),
				"int":               cty.NumberIntVal(int64(1)),
				"int64":             cty.NumberIntVal(int64(2)),
				"bool":              cty.True,
				"trilean":           cty.True,
				"duration":          cty.NumberIntVal((10 * time.Second).Milliseconds()),
				"map_string_string": cty.MapVal(map[string]cty.Value{"a": cty.StringVal("b")}),
				"slice_string": cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				}),
				"slice_slice_string": cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("b"),
				})}),
				"named_map_string_string": cty.MapVal(map[string]cty.Value{"a": cty.StringVal("b")}),
				"named_string":            cty.StringVal("named string"),
				"tag": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"key":   cty.StringVal("a"),
					"value": cty.StringVal("b"),
				})}),
				"data_source": cty.StringVal("datasource"),
				"nested": cty.ObjectVal(map[string]cty.Value{
					"string":            cty.StringVal("string"),
					"int":               cty.NumberIntVal(int64(1)),
					"int64":             cty.NumberIntVal(int64(2)),
					"bool":              cty.True,
					"trilean":           cty.True,
					"duration":          cty.NumberIntVal((10 * time.Second).Milliseconds()),
					"map_string_string": cty.MapVal(map[string]cty.Value{"a": cty.StringVal("b")}),
					"slice_string": cty.ListVal([]cty.Value{
						cty.StringVal("a"),
						cty.StringVal("b"),
					}),
					"slice_slice_string": cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{
						cty.StringVal("a"),
						cty.StringVal("b"),
					})}),
					"named_map_string_string": cty.MapVal(map[string]cty.Value{"a": cty.StringVal("b")}),
					"named_string":            cty.StringVal("named string"),
					"tag": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
						"key":   cty.StringVal("a"),
						"value": cty.StringVal("b"),
					})}),
					"data_source": cty.StringVal("datasource"),
				}),
				"nested_slice": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"string":            cty.StringVal("string"),
						"int":               cty.NumberIntVal(int64(1)),
						"int64":             cty.NumberIntVal(int64(2)),
						"bool":              cty.True,
						"trilean":           cty.True,
						"duration":          cty.NumberIntVal((10 * time.Second).Milliseconds()),
						"map_string_string": cty.MapVal(map[string]cty.Value{"a": cty.StringVal("b")}),
						"slice_string": cty.ListVal([]cty.Value{
							cty.StringVal("a"),
							cty.StringVal("b"),
						}),
						"slice_slice_string": cty.ListVal([]cty.Value{cty.ListVal([]cty.Value{
							cty.StringVal("a"),
							cty.StringVal("b"),
						})}),
						"named_map_string_string": cty.MapVal(map[string]cty.Value{"a": cty.StringVal("b")}),
						"named_string":            cty.StringVal("named string"),
						"tag": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
							"key":   cty.StringVal("a"),
							"value": cty.StringVal("b"),
						})}),
						"data_source": cty.StringVal("datasource"),
					}),
				}),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := HCL2ValueFromConfig(test.Input, test.Spec)
			if !reflect.DeepEqual(got, test.Want) {
				t.Errorf("wrong result\ninput: %#v\ngot:   %#v\nwant:  %#v", test.Input, got, test.Want)
			}
		})
	}
}
