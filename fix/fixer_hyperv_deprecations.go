// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerHypervDeprecations removes the deprecated "vhd_temp_path" setting
// from Hyper-V ISO builder templates
type FixerHypervDeprecations struct{}

func (FixerHypervDeprecations) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"MSOpenTech.hyperv": []string{"vhd_temp_path"},
	}
}

func (FixerHypervDeprecations) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if builderType != "hyperv-iso" {
			continue
		}

		_, ok = builder["vhd_temp_path"]
		if ok {
			delete(builder, "vhd_temp_path")
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerHypervDeprecations) Synopsis() string {
	return `Removes the deprecated "vhd_temp_path" setting from Hyper-V ISO builder templates`
}
