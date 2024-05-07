// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	commontpl "github.com/hashicorp/packer-plugin-sdk/template"
)

// ConsulFunc constructs a function that retrieves KV secrets from HC vault
var ConsulFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "key",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		key := args[0].AsString()
		val, err := commontpl.Consul(key)

		return cty.StringVal(val), err
	},
})
