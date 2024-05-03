// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/zclconf/go-cty/cty"
)

// HCL template usage example:
//
// locals {
//   emptyformat  = legacy_isotime()
//   golangformat = legacy_isotime("01-02-2006")
// }

func TestLegacyIsotime_empty(t *testing.T) {
	got, err := LegacyIsotimeFunc.Call([]cty.Value{})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = time.Parse(time.RFC3339, got.AsString())
	if err != nil {
		t.Fatalf("Didn't get an RFC3339 string from empty case: %s", err)
	}

}

func TestLegacyIsotime_inputs(t *testing.T) {
	tests := []struct {
		Value cty.Value
		Want  string
	}{
		{
			cty.StringVal("01-02-2006"),
			`^\d{2}-\d{2}-\d{4}$`,
		},
		{
			cty.StringVal("Mon Jan 02, 2006"),
			`^(Mon|Tue|Wed|Thu|Fri|Sat|Sun){1} (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec){1} \d{2}, \d{4}$`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("legacy_isotime(%#v)", test.Value), func(t *testing.T) {
			got, err := LegacyIsotime(test.Value)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			re, err := regexp.Compile(test.Want)
			if err != nil {
				t.Fatalf("Bad regular expression test string: %#v", err)
			}

			if !re.MatchString(got.AsString()) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
