// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"fmt"
	"time"

	"github.com/jehiah/go-strftime"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// initTime is the UTC time when this package was initialized. It is
// used as the timestamp for all configuration templates so that they
// match for a single build.
var initTime time.Time

func init() {
	initTime = time.Now().UTC()
}

// TimestampFunc constructs a function that returns a string representation of the current date and time.
var TimestampFunc = function.New(&function.Spec{
	Params:       []function.Parameter{},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(time.Now().UTC().Format(time.RFC3339)), nil
	},
})

// Timestamp returns a string representation of the current date and time.
//
// In the HCL language, timestamps are conventionally represented as strings
// using RFC 3339 "Date and Time format" syntax, and so timestamp returns a
// string in this format.
func Timestamp() (cty.Value, error) {
	return TimestampFunc.Call([]cty.Value{})
}

// LegacyIsotimeFunc constructs a function that returns a string representation
// of the current date and time using the Go language datetime formatting syntax.
var LegacyIsotimeFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name: "format",
		Type: cty.String,
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if len(args) > 1 {
			return cty.StringVal(""), fmt.Errorf("too many values, 1 needed: %v", args)
		}
		if len(args) == 0 {
			return cty.StringVal(initTime.Format(time.RFC3339)), nil
		}
		format := args[0].AsString()
		return cty.StringVal(initTime.Format(format)), nil
	},
})

// LegacyStrftimeFunc constructs a function that returns a string representation
// of the current date and time using the Go language strftime datetime formatting syntax.
var LegacyStrftimeFunc = function.New(&function.Spec{
	Params: []function.Parameter{},
	VarParam: &function.Parameter{
		Name: "format",
		Type: cty.String,
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		if len(args) > 1 {
			return cty.StringVal(""), fmt.Errorf("too many values, 1 needed: %v", args)
		}
		if len(args) == 0 {
			return cty.StringVal(initTime.Format(time.RFC3339)), nil
		}
		format := args[0].AsString()
		return cty.StringVal(strftime.Format(format, initTime)), nil
	},
})

// LegacyIsotimeFunc returns a string representation of the current date and
// time using the given format string. The format string follows golang's
// datetime formatting. See
// https://www.packer.io/docs/templates/legacy_json_templates/engine#isotime-function-format-reference
// for more details.
//
// This function has been provided to create backwards compatability with
// Packer's legacy JSON templates. However, we recommend that you upgrade your
// HCL Packer template to use Timestamp and FormatDate together as soon as is
// convenient.
//
// Please note that if you are using a large number of builders, provisioners
// or post-processors, the isotime may be slightly different for each one
// because it is from when the plugin is launched not the initial Packer
// process. In order to avoid this and make the timestamp consistent across all
// plugins, set it as a user variable and then access the user variable within
// your plugins.
func LegacyIsotime(format cty.Value) (cty.Value, error) {
	return LegacyIsotimeFunc.Call([]cty.Value{format})
}
