package function

import (
	commontpl "github.com/hashicorp/packer/common/template"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var AzureSecretFromKvId = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:         "key_vault_id",
			Type:         cty.String,
			AllowNull:    false,
			AllowUnknown: false,
		},
		{
			Name:         "key",
			Type:         cty.String,
			AllowNull:    false,
			AllowUnknown: false,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		keyVaultId := args[0].AsString()
		key := args[1].AsString()
		val, err := commontpl.GetAzureSecretFromKeyVaultId(keyVaultId, key)

		return cty.StringVal(val), err
	},
})

var AzureSecretFromKvRgAndName = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:         "resource_group_name",
			Type:         cty.String,
			AllowNull:    false,
			AllowUnknown: false,
		},
		{
			Name:         "key_vault_name",
			Type:         cty.String,
			AllowNull:    false,
			AllowUnknown: false,
		},
		{
			Name:         "key",
			Type:         cty.String,
			AllowNull:    false,
			AllowUnknown: false,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		resourceGroup := args[0].AsString()
		keyVaultName := args[1].AsString()
		key := args[2].AsString()
		val, err := commontpl.GetAzureSecretFromResourceGroupAndKeyVaultName(resourceGroup, keyVaultName, key)

		return cty.StringVal(val), err
	},
})
