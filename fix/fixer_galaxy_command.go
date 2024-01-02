// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerGalaxyCommand removes the escape character from user
// environment variables and replace galaxycommand with galaxy_command
type FixerGalaxyCommand struct{}

func (FixerGalaxyCommand) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"ansible": []string{"galaxycommand"},
	}
}

func (FixerGalaxyCommand) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	type template struct {
		Provisioners []interface{}
	}

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

		if ok := provisioners["type"] == "ansible-local"; !ok {
			continue
		}

		if _, ok := provisioners["galaxy_command"]; ok {

			// drop galaxycommand if it is also included
			if _, galaxyCommandIncluded := provisioners["galaxycommand"]; galaxyCommandIncluded {
				delete(provisioners, "galaxycommand")
			}

		} else {

			// replace galaxycommand with galaxy_command if it exists
			galaxyCommandRaw, ok := provisioners["galaxycommand"]
			if !ok {
				continue
			}

			galaxyCommandString, ok := galaxyCommandRaw.(string)
			if !ok {
				continue
			}

			delete(provisioners, "galaxycommand")
			provisioners["galaxy_command"] = galaxyCommandString
		}

		// Write all changes back to template
		tpl.Provisioners[i] = provisioners
	}

	if len(tpl.Provisioners) > 0 {
		input["provisioners"] = tpl.Provisioners
	}

	return input, nil
}

func (FixerGalaxyCommand) Synopsis() string {
	return `Replaces "galaxycommand" in ansible-local provisioner configs with "galaxy_command"`
}
