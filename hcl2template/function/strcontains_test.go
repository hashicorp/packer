// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestStrContains(t *testing.T) {
	tests := []struct {
		String    cty.Value
		Substr    cty.Value
		Want      cty.Value
		ExpectErr bool
	}{
		{
			cty.StringVal("hello"),
			cty.StringVal("hel"),
			cty.BoolVal(true),
			false,
		},
		{
			cty.StringVal("hello"),
			cty.StringVal("lo"),
			cty.BoolVal(true),
			false,
		},
		{
			cty.StringVal("hello1"),
			cty.StringVal("1"),
			cty.BoolVal(true),
			false,
		},
		{
			cty.StringVal("hello1"),
			cty.StringVal("heo"),
			cty.BoolVal(false),
			false,
		},
		{
			cty.StringVal("hello1"),
			cty.NumberIntVal(1),
			cty.UnknownVal(cty.Bool),
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("includes(%#v, %#v)", test.String, test.Substr), func(t *testing.T) {
			got, err := StrContains.Call([]cty.Value{
				test.String,
				test.Substr,
			})

			if test.ExpectErr && err == nil {
				t.Fatal("succeeded; want error")
			}

			if test.ExpectErr && err != nil {
				return
			}

			if !test.ExpectErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
