package function

import (
	"fmt"
	"math"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestSum(t *testing.T) {
	tests := []struct {
		List cty.Value
		Want cty.Value
		Err  string
	}{
		{
			cty.ListVal([]cty.Value{
				cty.NumberIntVal(1),
				cty.NumberIntVal(2),
				cty.NumberIntVal(3),
			}),
			cty.NumberIntVal(6),
			"",
		},
		{
			cty.ListVal([]cty.Value{
				cty.NumberIntVal(1476),
				cty.NumberIntVal(2093),
				cty.NumberIntVal(2092495),
				cty.NumberIntVal(64589234),
				cty.NumberIntVal(234),
			}),
			cty.NumberIntVal(66685532),
			"",
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
			cty.UnknownVal(cty.String),
			"argument must be list, set, or tuple of number values",
		},
		{
			cty.ListVal([]cty.Value{
				cty.NumberIntVal(10),
				cty.NumberIntVal(-19),
				cty.NumberIntVal(5),
			}),
			cty.NumberIntVal(-4),
			"",
		},
		{
			cty.ListVal([]cty.Value{
				cty.NumberFloatVal(10.2),
				cty.NumberFloatVal(19.4),
				cty.NumberFloatVal(5.7),
			}),
			cty.NumberFloatVal(35.3),
			"",
		},
		{
			cty.ListVal([]cty.Value{
				cty.NumberFloatVal(-10.2),
				cty.NumberFloatVal(-19.4),
				cty.NumberFloatVal(-5.7),
			}),
			cty.NumberFloatVal(-35.3),
			"",
		},
		{
			cty.ListVal([]cty.Value{cty.NullVal(cty.Number)}),
			cty.NilVal,
			"argument must be list, set, or tuple of number values",
		},
		{
			cty.ListVal([]cty.Value{
				cty.NumberIntVal(5),
				cty.NullVal(cty.Number),
			}),
			cty.NilVal,
			"argument must be list, set, or tuple of number values",
		},
		{
			cty.SetVal([]cty.Value{
				cty.StringVal("a"),
				cty.StringVal("b"),
				cty.StringVal("c"),
			}),
			cty.UnknownVal(cty.String).RefineNotNull(),
			"argument must be list, set, or tuple of number values",
		},
		{
			cty.SetVal([]cty.Value{
				cty.NumberIntVal(10),
				cty.NumberIntVal(-19),
				cty.NumberIntVal(5),
			}),
			cty.NumberIntVal(-4),
			"",
		},
		{
			cty.SetVal([]cty.Value{
				cty.NumberIntVal(10),
				cty.NumberIntVal(25),
				cty.NumberIntVal(30),
			}),
			cty.NumberIntVal(65),
			"",
		},
		{
			cty.SetVal([]cty.Value{
				cty.NumberFloatVal(2340.8),
				cty.NumberFloatVal(10.2),
				cty.NumberFloatVal(3),
			}),
			cty.NumberFloatVal(2354),
			"",
		},
		{
			cty.SetVal([]cty.Value{
				cty.NumberFloatVal(2),
			}),
			cty.NumberFloatVal(2),
			"",
		},
		{
			cty.SetVal([]cty.Value{
				cty.NumberFloatVal(-2),
				cty.NumberFloatVal(-50),
				cty.NumberFloatVal(-20),
				cty.NumberFloatVal(-123),
				cty.NumberFloatVal(-4),
			}),
			cty.NumberFloatVal(-199),
			"",
		},
		{
			cty.TupleVal([]cty.Value{
				cty.NumberIntVal(12),
				cty.StringVal("a"),
				cty.NumberIntVal(38),
			}),
			cty.UnknownVal(cty.String).RefineNotNull(),
			"argument must be list, set, or tuple of number values",
		},
		{
			cty.NumberIntVal(12),
			cty.NilVal,
			"cannot sum noniterable",
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.NilVal,
			"cannot sum an empty list",
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.NilVal,
			"argument must be list, set, or tuple. Received map of bool",
		},
		{
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number).RefineNotNull(),
			"",
		},
		{
			cty.UnknownVal(cty.List(cty.Number)),
			cty.UnknownVal(cty.Number).RefineNotNull(),
			"",
		},
		{ // known list containing unknown values
			cty.ListVal([]cty.Value{cty.UnknownVal(cty.Number)}),
			cty.UnknownVal(cty.Number).RefineNotNull(),
			"",
		},
		{ // numbers too large to represent as float64
			cty.ListVal([]cty.Value{
				cty.MustParseNumberVal("1e+500"),
				cty.MustParseNumberVal("1e+500"),
			}),
			cty.MustParseNumberVal("2e+500"),
			"",
		},
		{ // edge case we have a special error handler for
			cty.ListVal([]cty.Value{
				cty.NumberFloatVal(math.Inf(1)),
				cty.NumberFloatVal(math.Inf(-1)),
			}),
			cty.NilVal,
			"can't compute sum of opposing infinities",
		},
		{
			cty.ListVal([]cty.Value{
				cty.StringVal("1"),
				cty.StringVal("2"),
				cty.StringVal("3"),
			}),
			cty.NumberIntVal(6),
			"",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sum(%#v)", test.List), func(t *testing.T) {
			got, err := Sum(test.List)

			if test.Err != "" {
				if err == nil {
					t.Fatal("succeeded; want error")
				} else if got, want := err.Error(), test.Err; got != want {
					t.Fatalf("wrong error\n got: %s\nwant: %s", got, want)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
