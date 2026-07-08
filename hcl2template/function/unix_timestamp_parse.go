// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"time"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

// UnixTimestampParseFunc constructs a function that parses a Unix timestamp
// integer (the number of seconds elapsed since January 1, 1970 UTC) and returns
// an object representation of that date and time, including its RFC 3339 form.
var UnixTimestampParseFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "unix_timestamp",
			Type: cty.Number,
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
		"rfc3339":      cty.String,
		"second":       cty.Number,
		"iso_year":     cty.Number,
		"iso_week":     cty.Number,
	})),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		// Propagate unknown values rather than attempting a conversion, which
		// would fail for an unknown input.
		if !args[0].IsKnown() {
			return cty.UnknownVal(retType), nil
		}

		var ts int64
		if err := gocty.FromCtyValue(args[0], &ts); err != nil {
			return cty.NilVal, function.NewArgErrorf(0, "unix_timestamp must be a whole number of seconds representable as a 64-bit integer: %s", err)
		}

		parsed := time.Unix(ts, 0).UTC()
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
			"rfc3339":      cty.StringVal(parsed.Format(time.RFC3339)),
			"second":       cty.NumberIntVal(int64(parsed.Second())),
			"iso_year":     cty.NumberIntVal(int64(isoYear)),
			"iso_week":     cty.NumberIntVal(int64(isoWeek)),
		}), nil
	},
})

// UnixTimestampParse parses a Unix timestamp integer and returns an object
// representation of that date and time, including its RFC 3339 form in the
// "rfc3339" attribute.
func UnixTimestampParse(unixTimestamp cty.Value) (cty.Value, error) {
	return UnixTimestampParseFunc.Call([]cty.Value{unixTimestamp})
}
