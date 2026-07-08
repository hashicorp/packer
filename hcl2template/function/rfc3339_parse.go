// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"time"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// rfc3339Layouts are the timestamp layouts accepted by parseRFC3339. RFC 3339
// permits either an uppercase "T" or a space to separate the date and time
// (section 5.6), and an optional fractional-seconds component. The Nano layouts
// make the fractional part optional, so these two layouts together accept the
// full range of RFC 3339 timestamps regardless of separator or fractional
// seconds.
var rfc3339Layouts = []string{
	time.RFC3339Nano,                      // "T" separator, e.g. 2023-07-25T23:43:16.5Z
	"2006-01-02 15:04:05.999999999Z07:00", // space separator, e.g. 2023-07-25 23:43:16.5Z
}

// parseRFC3339 parses an RFC 3339 timestamp, accepting either separator and an
// optional fractional-seconds component.
func parseRFC3339(ts string) (time.Time, bool) {
	for _, layout := range rfc3339Layouts {
		if parsed, err := time.Parse(layout, ts); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

// RFC3339ParseFunc constructs a function that parses an RFC 3339 timestamp
// string and returns an object representation of that date and time, including
// its Unix timestamp form.
var RFC3339ParseFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "timestamp",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.Object(map[string]cty.Type{
		"year":         cty.Number,
		"year_day":     cty.Number,
		"day":          cty.Number,
		"month":        cty.Number,
		"month_name":   cty.String,
		"weekday":      cty.Number,
		"weekday_name": cty.String,
		"hour":         cty.Number,
		"minute":       cty.Number,
		"second":       cty.Number,
		"unix":         cty.Number,
		"iso_year":     cty.Number,
		"iso_week":     cty.Number,
	})),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if !args[0].IsKnown() {
			return cty.UnknownVal(retType), nil
		}

		ts := args[0].AsString()

		parsed, ok := parseRFC3339(ts)
		if !ok {
			// Intentionally omit the underlying Go parse error, as its message
			// references Go's reference time and is confusing to template
			// authors.
			return cty.NilVal, function.NewArgErrorf(0, "%q is not a valid RFC3339 timestamp", ts)
		}

		isoYear, isoWeek := parsed.ISOWeek()

		return cty.ObjectVal(map[string]cty.Value{
			"year":         cty.NumberIntVal(int64(parsed.Year())),
			"year_day":     cty.NumberIntVal(int64(parsed.YearDay())),
			"day":          cty.NumberIntVal(int64(parsed.Day())),
			"month":        cty.NumberIntVal(int64(parsed.Month())),
			"month_name":   cty.StringVal(parsed.Month().String()),
			"weekday":      cty.NumberIntVal(int64(parsed.Weekday())),
			"weekday_name": cty.StringVal(parsed.Weekday().String()),
			"hour":         cty.NumberIntVal(int64(parsed.Hour())),
			"minute":       cty.NumberIntVal(int64(parsed.Minute())),
			"second":       cty.NumberIntVal(int64(parsed.Second())),
			"unix":         cty.NumberIntVal(parsed.Unix()),
			"iso_year":     cty.NumberIntVal(int64(isoYear)),
			"iso_week":     cty.NumberIntVal(int64(isoWeek)),
		}), nil
	},
})

// RFC3339Parse parses an RFC 3339 timestamp string and returns an object
// representation of that date and time, including its Unix timestamp form in
// the "unix" attribute.
func RFC3339Parse(timestamp cty.Value) (cty.Value, error) {
	return RFC3339ParseFunc.Call([]cty.Value{timestamp})
}
