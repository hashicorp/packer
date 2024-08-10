// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestBase64Gzip(t *testing.T) {
	tests := []struct {
		String cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			cty.StringVal("test"),
			cty.StringVal("H4sIAAAAAAAA/ypJLS4BAAAA//8BAAD//wx+f9gEAAAA"),
			false,
		},
		{
			cty.StringVal("helloworld"),
			cty.StringVal("H4sIAAAAAAAA/8pIzcnJL88vykkBAAAA//8BAAD//60g6/kKAAAA"),
			false,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("base64gzip(%#v)", test.String), func(t *testing.T) {
			got, err := Base64Gzip(test.String)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
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
