package function

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestEndsWith(t *testing.T) {
	tests := []struct {
		String, Suffix cty.Value
		Want           cty.Value
	}{
		{
			cty.StringVal("hello world"),
			cty.StringVal("world"),
			cty.True,
		},
		{
			cty.StringVal("hey world"),
			cty.StringVal("worldss"),
			cty.False,
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
			cty.True,
		},
		{
			cty.StringVal("a"),
			cty.StringVal(""),
			cty.True,
		},
		{
			cty.StringVal("hello world"),
			cty.StringVal(" "),
			cty.False,
		},
		{
			cty.StringVal(" "),
			cty.StringVal(""),
			cty.True,
		},
		{
			cty.StringVal(" "),
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.StringVal(""),
			cty.StringVal("a"),
			cty.False,
		},
		{
			cty.UnknownVal(cty.String),
			cty.StringVal("a"),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
		{
			cty.UnknownVal(cty.String),
			cty.StringVal(""),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("EndsWith(%#v, %#v)", test.String, test.Suffix), func(t *testing.T) {
			got, err := EndsWithFunc.Call([]cty.Value{test.String, test.Suffix})

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf(
					"wrong result\nstring: %#v\nsuffix: %#v\ngot:    %#v\nwant:   %#v",
					test.String, test.Suffix, got, test.Want,
				)
			}
		})
	}
}
