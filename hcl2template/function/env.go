// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package function

import (
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// EnvFunc constructs a function that returns a string representation of the
// env var behind a value
var EnvFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:         "key",
			Type:         cty.String,
			AllowNull:    false,
			AllowUnknown: false,
		},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		key := args[0].AsString()
		value := os.Getenv(key)
		return cty.StringVal(value), nil
	},
})

// Env returns a string representation of the env var behind key.
func Env(key cty.Value) (cty.Value, error) {
	return EnvFunc.Call([]cty.Value{key})
}
