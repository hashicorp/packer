// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
)

func testVar(name string, typ cty.Type, val cty.Value, sensitive bool) *Variable {
	return &Variable{
		Name:      name,
		Type:      typ,
		Sensitive: sensitive,
		Values: []VariableAssignment{
			{From: "default", Value: val},
		},
	}
}

func TestPackerConfig_userVariableValues(t *testing.T) {
	cfg := &PackerConfig{
		InputVariables: Variables{
			"str":    testVar("str", cty.String, cty.StringVal("hello"), false),
			"flag":   testVar("flag", cty.Bool, cty.True, false),
			"num":    testVar("num", cty.Number, cty.NumberIntVal(42), false),
			"nul":    testVar("nul", cty.String, cty.NullVal(cty.String), false),
			"secret": testVar("secret", cty.String, cty.StringVal("hunter2"), true),
			"list": testVar("list", cty.List(cty.String),
				cty.ListVal([]cty.Value{cty.StringVal("a")}), false),
			"unset": {Name: "unset", Type: cty.String},
		},
	}

	got := cfg.userVariableValues()
	want := map[string]string{
		"str":  "hello",
		"flag": "true",
		"num":  "42",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("userVariableValues() mismatch (-want +got):\n%s", diff)
	}
}
