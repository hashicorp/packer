// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVirtualBoxRename changes "virtualbox" builders to "virtualbox-iso"
type FixerVirtualBoxRename struct{}

func (FixerVirtualBoxRename) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerVirtualBoxRename) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	type template struct {
		Builders     []map[string]interface{}
		Provisioners []interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.WeakDecode(input, &tpl); err != nil {
		return nil, err
	}

	for _, builder := range tpl.Builders {
		builderTypeRaw, ok := builder["type"]
		if !ok {
			continue
		}

		builderType, ok := builderTypeRaw.(string)
		if !ok {
			continue
		}

		if builderType != "virtualbox" {
			continue
		}

		builder["type"] = "virtualbox-iso"
	}

	for i, raw := range tpl.Provisioners {
		var m map[string]interface{}
		if err := mapstructure.WeakDecode(raw, &m); err != nil {
			// Ignore errors, could be a non-map
			continue
		}

		raw, ok := m["override"]
		if !ok {
			continue
		}

		var override map[string]interface{}
		if err := mapstructure.WeakDecode(raw, &override); err != nil {
			return nil, err
		}

		if raw, ok := override["virtualbox"]; ok {
			override["virtualbox-iso"] = raw
			delete(override, "virtualbox")

			// Set the change
			m["override"] = override
			tpl.Provisioners[i] = m
		}
	}

	if len(tpl.Builders) > 0 {
		input["builders"] = tpl.Builders
	}
	if len(tpl.Provisioners) > 0 {
		input["provisioners"] = tpl.Provisioners
	}
	return input, nil
}

func (FixerVirtualBoxRename) Synopsis() string {
	return `Updates "virtualbox" builders to "virtualbox-iso"`
}
