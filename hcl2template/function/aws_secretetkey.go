// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	commontpl "github.com/hashicorp/packer-plugin-sdk/template"
)

// AWSSecret constructs a function that retrieves secrets from aws secrets
// manager. If Key field is not set then we will return first secret key stored
// in secret name.
var AWSSecret = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:         "name",
			Type:         cty.String,
			AllowNull:    false,
			AllowUnknown: false,
		},
		{
			Name:         "key",
			Type:         cty.String,
			AllowNull:    true,
			AllowUnknown: false,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		name := args[0].AsString()
		var key string
		if !args[1].IsNull() && args[1].IsWhollyKnown() {
			key = args[1].AsString()
		}
		val, err := commontpl.GetAWSSecret(name, key)

		return cty.StringVal(val), err
	},
})
