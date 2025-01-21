// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestAnyTrue(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Want       cty.Value
		Err        bool
	}{
		{
			cty.ListValEmpty(cty.Bool),
			cty.False,
			false,
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.True,
			false,
		},
		{
			cty.ListVal([]cty.Value{cty.False}),
			cty.False,
			false,
		},
		{
			cty.ListVal([]cty.Value{cty.True, cty.False}),
			cty.True,
			false,
		},
		{
			cty.ListVal([]cty.Value{cty.False, cty.True}),
			cty.True,
			false,
		},
		{
			cty.ListVal([]cty.Value{cty.True, cty.NullVal(cty.Bool)}),
			cty.True,
			false,
		},
		{
			cty.ListVal([]cty.Value{cty.UnknownVal(cty.Bool)}),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
			false,
		},
		{
			cty.ListVal([]cty.Value{
				cty.UnknownVal(cty.Bool),
				cty.UnknownVal(cty.Bool),
			}),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
			false,
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Bool).RefineNotNull(),
			false,
		},
		{
			cty.NullVal(cty.List(cty.Bool)),
			cty.NilVal,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("anytrue(%#v)", tc.Collection), func(t *testing.T) {
			got, err := AnyTrue.Call([]cty.Value{tc.Collection})

			if tc.Err && err == nil {
				t.Fatal("succeeded; want error")
			}
			if !tc.Err && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !got.RawEquals(tc.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, tc.Want)
			}
		})
	}
}
