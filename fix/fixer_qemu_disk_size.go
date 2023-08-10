// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"strconv"

	"github.com/mitchellh/mapstructure"
)

// FixerQEMUDiskSize updates disk_size from a string to int for QEMU builders
type FixerQEMUDiskSize struct{}

func (FixerQEMUDiskSize) DeprecatedOptions() map[string][]string {
	return map[string][]string{}
}

func (FixerQEMUDiskSize) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if builderType != "qemu" {
			continue
		}

		switch diskSize := builder["disk_size"].(type) {
		case float64:
			builder["disk_size"] = strconv.Itoa(int(diskSize)) + "M"
		case int:
			builder["disk_size"] = strconv.Itoa(diskSize) + "M"
		}

	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerQEMUDiskSize) Synopsis() string {
	return `Updates "disk_size" from int to string in QEMU builders.`
}
