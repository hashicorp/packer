// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

// HCL template usage example:
//
// locals {
//   parsed  = unix_timestamp_parse(1690328596)
//   rfc3339 = local.parsed.rfc3339
// }

func TestUnixTimestampParse(t *testing.T) {
	tests := []struct {
		Value cty.Value
		Want  cty.Value
	}{
		{
			cty.NumberIntVal(1690328596),
			cty.ObjectVal(map[string]cty.Value{
				"year":         cty.NumberIntVal(2023),
				"year_day":     cty.NumberIntVal(206),
				"day":          cty.NumberIntVal(25),
				"month":        cty.NumberIntVal(7),
				"month_name":   cty.StringVal("July"),
				"weekday":      cty.NumberIntVal(2),
				"weekday_name": cty.StringVal("Tuesday"),
				"hour":         cty.NumberIntVal(23),
				"minute":       cty.NumberIntVal(43),
				"rfc3339":      cty.StringVal("2023-07-25T23:43:16Z"),
				"second":       cty.NumberIntVal(16),
				"iso_year":     cty.NumberIntVal(2023),
				"iso_week":     cty.NumberIntVal(30),
			}),
		},
		{
			cty.NumberIntVal(851042397),
			cty.ObjectVal(map[string]cty.Value{
				"year":         cty.NumberIntVal(1996),
				"year_day":     cty.NumberIntVal(355),
				"day":          cty.NumberIntVal(20),
				"month":        cty.NumberIntVal(12),
				"month_name":   cty.StringVal("December"),
				"weekday":      cty.NumberIntVal(5),
				"weekday_name": cty.StringVal("Friday"),
				"hour":         cty.NumberIntVal(0),
				"minute":       cty.NumberIntVal(39),
				"rfc3339":      cty.StringVal("1996-12-20T00:39:57Z"),
				"second":       cty.NumberIntVal(57),
				"iso_year":     cty.NumberIntVal(1996),
				"iso_week":     cty.NumberIntVal(51),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.AsBigFloat().Text('f', 0), func(t *testing.T) {
			got, err := UnixTimestampParse(test.Value)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestUnixTimestampParse_invalid(t *testing.T) {
	tests := []struct {
		Name  string
		Value cty.Value
	}{
		{
			// Unix timestamps must be whole seconds; a fractional value
			// cannot be represented as an int64 and must be rejected.
			"fractional",
			cty.NumberFloatVal(1690328596.5),
		},
		{
			// Values outside the range of int64 cannot be represented.
			"out_of_range",
			cty.MustParseNumberVal("100000000000000000000"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := UnixTimestampParse(test.Value)
			if err == nil {
				t.Fatalf("expected error for %s input, got none", test.Name)
			}
		})
	}
}

func TestUnixTimestampParse_unknown(t *testing.T) {
	got, err := UnixTimestampParse(cty.UnknownVal(cty.Number))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got.IsKnown() {
		t.Errorf("expected unknown result for unknown input, got: %#v", got)
	}
}
