// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerVMwareRename changes "vmware" builders to "vmware-iso"
type FixerVMwareRename struct{}

func (FixerVMwareRename) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerVMwareRename) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// The type we'll decode into; we only care about builders
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

		if builderType != "vmware" {
			continue
		}

		builder["type"] = "vmware-iso"
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerVMwareRename) Synopsis() string {
	return `Updates "vmware" builders to "vmware-iso"`
}
