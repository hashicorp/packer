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
//   parsed = rfc3339_parse("2023-07-25T23:43:16Z")
//   epoch  = local.parsed.unix
// }

func TestRFC3339Parse(t *testing.T) {
	tests := []struct {
		Value cty.Value
		Want  cty.Value
	}{
		{
			cty.StringVal("2023-07-25T23:43:16Z"),
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
				"second":       cty.NumberIntVal(16),
				"unix":         cty.NumberIntVal(1690328596),
				"iso_year":     cty.NumberIntVal(2023),
				"iso_week":     cty.NumberIntVal(30),
			}),
		},
		{
			// Offsets are normalized to UTC when computing the Unix timestamp.
			cty.StringVal("1996-12-19T16:39:57-08:00"),
			cty.ObjectVal(map[string]cty.Value{
				"year":         cty.NumberIntVal(1996),
				"year_day":     cty.NumberIntVal(354),
				"day":          cty.NumberIntVal(19),
				"month":        cty.NumberIntVal(12),
				"month_name":   cty.StringVal("December"),
				"weekday":      cty.NumberIntVal(4),
				"weekday_name": cty.StringVal("Thursday"),
				"hour":         cty.NumberIntVal(16),
				"minute":       cty.NumberIntVal(39),
				"second":       cty.NumberIntVal(57),
				"unix":         cty.NumberIntVal(851042397),
				"iso_year":     cty.NumberIntVal(1996),
				"iso_week":     cty.NumberIntVal(51),
			}),
		},
		{
			// RFC 3339 section 5.6 permits a space in place of "T" as the
			// date/time separator. This parses to the same instant as the
			// first case.
			cty.StringVal("2023-07-25 23:43:16Z"),
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
				"second":       cty.NumberIntVal(16),
				"unix":         cty.NumberIntVal(1690328596),
				"iso_year":     cty.NumberIntVal(2023),
				"iso_week":     cty.NumberIntVal(30),
			}),
		},
		{
			// RFC 3339 permits fractional seconds. The fractional component is
			// dropped from the whole-second "second" and "unix" attributes.
			cty.StringVal("2023-07-25T23:43:16.512345Z"),
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
				"second":       cty.NumberIntVal(16),
				"unix":         cty.NumberIntVal(1690328596),
				"iso_year":     cty.NumberIntVal(2023),
				"iso_week":     cty.NumberIntVal(30),
			}),
		},
		{
			// Space separator combined with fractional seconds and an offset.
			cty.StringVal("1996-12-19 16:39:57.25-08:00"),
			cty.ObjectVal(map[string]cty.Value{
				"year":         cty.NumberIntVal(1996),
				"year_day":     cty.NumberIntVal(354),
				"day":          cty.NumberIntVal(19),
				"month":        cty.NumberIntVal(12),
				"month_name":   cty.StringVal("December"),
				"weekday":      cty.NumberIntVal(4),
				"weekday_name": cty.StringVal("Thursday"),
				"hour":         cty.NumberIntVal(16),
				"minute":       cty.NumberIntVal(39),
				"second":       cty.NumberIntVal(57),
				"unix":         cty.NumberIntVal(851042397),
				"iso_year":     cty.NumberIntVal(1996),
				"iso_week":     cty.NumberIntVal(51),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.Value.AsString(), func(t *testing.T) {
			got, err := RFC3339Parse(test.Value)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestRFC3339Parse_invalid(t *testing.T) {
	_, err := RFC3339Parse(cty.StringVal("abcdef"))
	if err == nil {
		t.Fatalf("expected error for invalid timestamp, got none")
	}

	want := `"abcdef" is not a valid RFC3339 timestamp`
	if got := err.Error(); got != want {
		t.Errorf("wrong error\ngot:  %s\nwant: %s", got, want)
	}
}

func TestRFC3339Parse_unknown(t *testing.T) {
	got, err := RFC3339Parse(cty.UnknownVal(cty.String))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got.IsKnown() {
		t.Errorf("expected unknown result for unknown input, got: %#v", got)
	}
}
