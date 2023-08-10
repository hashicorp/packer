// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerProxmoxType updates proxmox builder types to proxmox-iso
type FixerProxmoxType struct{}

func (FixerProxmoxType) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerProxmoxType) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
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

		if builderType != "proxmox" {
			continue
		}

		builder["type"] = "proxmox-iso"
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerProxmoxType) Synopsis() string {
	return `Updates the builder type proxmox to proxmox-iso`
}
