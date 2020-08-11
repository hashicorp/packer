package function

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	vaultapi "github.com/hashicorp/vault/api"
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

		if token := os.Getenv("VAULT_TOKEN"); token == "" {
			return cty.StringVal(""), errors.New("Must set VAULT_TOKEN env var in order to " +
				"use vault template function")
		}

		vaultConfig := vaultapi.DefaultConfig()
		cli, err := vaultapi.NewClient(vaultConfig)
		if err != nil {
			return cty.StringVal(""), fmt.Errorf("Error getting Vault client: %s", err)
		}
		secret, err := cli.Logical().Read(path)
		if err != nil {
			return cty.StringVal(""), fmt.Errorf("Error reading vault secret: %s", err)
		}
		if secret == nil {
			return cty.StringVal(""), errors.New("Vault Secret does not exist at the given path")
		}

		data, ok := secret.Data["data"]
		if !ok {
			// maybe ths is v1, not v2 kv store
			value, ok := secret.Data[key]
			if ok {
				return cty.StringVal(value.(string)), nil
			}

			// neither v1 nor v2 proudced a valid value
			return cty.StringVal(""), fmt.Errorf("Vault data was empty at the "+
				"given path. Warnings: %s", strings.Join(secret.Warnings, "; "))
		}

		value := data.(map[string]interface{})[key].(string)
		return cty.StringVal(value), nil
	},
})

// Vault returns a secret from a KV store in HC vault
func Vault() (cty.Value, error) {
	return VaultFunc.Call([]cty.Value{})
}
