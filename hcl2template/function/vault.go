// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	commontpl "github.com/hashicorp/packer-plugin-sdk/template"
)

// VaultFunc constructs a function that retrieves KV secrets from HC vault
var VaultFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
		{
			Name: "key",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		path := args[0].AsString()
		key := args[1].AsString()

		val, err := commontpl.Vault(path, key)

		return cty.StringVal(val), err
	},
})
