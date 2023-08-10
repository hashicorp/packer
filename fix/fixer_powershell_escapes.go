// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

// FixerPowerShellEscapes removes the PowerShell escape character from user
// environment variables and elevated username and password strings
type FixerPowerShellEscapes struct{}

func (FixerPowerShellEscapes) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerPowerShellEscapes) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	type template struct {
		Provisioners []interface{}
	}

	var psUnescape = strings.NewReplacer(
		"`$", "$",
		"`\"", "\"",
		"``", "`",
		"`'", "'",
	)

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.WeakDecode(input, &tpl); err != nil {
		return nil, err
	}

	for i, raw := range tpl.Provisioners {
		var provisioners map[string]interface{}
		if err := mapstructure.Decode(raw, &provisioners); err != nil {
			// Ignore errors, could be a non-map
			continue
		}

		if ok := provisioners["type"] == "powershell"; !ok {
			continue
		}

		if _, ok := provisioners["elevated_user"]; ok {
			provisioners["elevated_user"] = psUnescape.Replace(provisioners["elevated_user"].(string))
		}
		if _, ok := provisioners["elevated_password"]; ok {
			provisioners["elevated_password"] = psUnescape.Replace(provisioners["elevated_password"].(string))
		}
		if raw, ok := provisioners["environment_vars"]; ok {
			var env_vars []string
			if err := mapstructure.Decode(raw, &env_vars); err != nil {
				continue
			}
			env_vars_unescaped := make([]interface{}, len(env_vars))
			for j, env_var := range env_vars {
				env_vars_unescaped[j] = psUnescape.Replace(env_var)
			}
			// Replace with unescaped environment variables
			provisioners["environment_vars"] = env_vars_unescaped
		}

		// Write all changes back to template
		tpl.Provisioners[i] = provisioners
	}

	if len(tpl.Provisioners) > 0 {
		input["provisioners"] = tpl.Provisioners
	}

	return input, nil
}

func (FixerPowerShellEscapes) Synopsis() string {
	return `Removes PowerShell escapes from user env vars and elevated username and password strings`
}
