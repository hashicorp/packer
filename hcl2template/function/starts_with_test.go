package function

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestStartsWith(t *testing.T) {
	tests := []struct {
		String, Prefix cty.Value
		Want           cty.Value
		WantError      string
	}{
		{
			cty.StringVal("hello world"),
			cty.StringVal("hello"),
			cty.True,
			``,
		},
		{
			cty.StringVal("hey world"),
			cty.StringVal("hello"),
			cty.False,
			``,
		},
		{
			cty.StringVal(""),
			cty.StringVal(""),
			cty.True,
			``,
		},
		{
			cty.StringVal("a"),
			cty.StringVal(""),
			cty.True,
			``,
		},
		{
			cty.StringVal(""),
			cty.StringVal("a"),
			cty.False,
			``,
		},
		{
			cty.UnknownVal(cty.String),
			cty.StringVal("a"),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
			``,
		},
		{
			cty.UnknownVal(cty.String),
			cty.StringVal(""),
			cty.True,
			``,
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("https:").NewValue(),
			cty.StringVal(""),
			cty.True,
			``,
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("https:").NewValue(),
			cty.StringVal("a"),
			cty.False,
			``,
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("https:").NewValue(),
			cty.StringVal("ht"),
			cty.True,
			``,
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("https:").NewValue(),
			cty.StringVal("https:"),
			cty.True,
			``,
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("https:").NewValue(),
			cty.StringVal("https-"),
			cty.False,
			``,
		},
		{
			cty.UnknownVal(cty.String).Refine().StringPrefix("https:").NewValue(),
			cty.StringVal("https://"),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
			``,
		},
		{
			// Unicode combining characters edge-case: we match the prefix
			// in terms of unicode code units rather than grapheme clusters,
			// which is inconsistent with our string processing elsewhere but
			// would be a breaking change to fix that bug now.
			cty.StringVal("\U0001f937\u200d\u2642"), // "Man Shrugging" is encoded as "Person Shrugging" followed by zero-width joiner and then the masculine gender presentation modifier
			cty.StringVal("\U0001f937"),             // Just the "Person Shrugging" character without any modifiers
			cty.True,
			``,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("StartsWith(%#v, %#v)", test.String, test.Prefix), func(t *testing.T) {
			got, err := StartsWithFunc.Call([]cty.Value{test.String, test.Prefix})

			if test.WantError != "" {
				gotErr := fmt.Sprintf("%s", err)
				if gotErr != test.WantError {
					t.Errorf("wrong error\ngot:  %s\nwant: %s", gotErr, test.WantError)
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf(
					"wrong result\nstring: %#v\nprefix: %#v\ngot:    %#v\nwant:   %#v",
					test.String, test.Prefix, got, test.Want,
				)
			}
		})
	}
}
