// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// Base64GzipFunc constructs a function that compresses a string with gzip and then encodes the result in
// Base64 encoding.
var Base64GzipFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		s := args[0].AsString()

		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		if _, err := gz.Write([]byte(s)); err != nil {
			return cty.UnknownVal(cty.String), fmt.Errorf("failed to write gzip raw data: %w", err)
		}
		if err := gz.Flush(); err != nil {
			return cty.UnknownVal(cty.String), fmt.Errorf("failed to flush gzip writer: %w", err)
		}
		if err := gz.Close(); err != nil {
			return cty.UnknownVal(cty.String), fmt.Errorf("failed to close gzip writer: %w", err)
		}
		return cty.StringVal(base64.StdEncoding.EncodeToString(b.Bytes())), nil
	},
})

// Base64Gzip compresses a string with gzip and then encodes the result in
// Base64 encoding.
//
// Packer uses the "standard" Base64 alphabet as defined in RFC 4648 section 4.
//
// Strings in the Packer language are sequences of unicode characters rather
// than bytes, so this function will first encode the characters from the string
// as UTF-8, then apply gzip compression, and then finally apply Base64 encoding.
func Base64Gzip(str cty.Value) (cty.Value, error) {
	return Base64GzipFunc.Call([]cty.Value{str})
}
