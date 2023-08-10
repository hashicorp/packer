// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerHypervVmxcTypo fixes the typo in "clone_from_vmxc_path" replacing
// it with "clone_from_vmcx_path" in Hyper-V VMCX builder templates
type FixerHypervVmxcTypo struct{}

func (FixerHypervVmxcTypo) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"MSOpenTech.hyperv": []string{"clone_from_vmxc_path"},
	}
}

func (FixerHypervVmxcTypo) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if builderType != "hyperv-vmcx" {
			continue
		}

		path, ok := builder["clone_from_vmxc_path"]
		if ok {
			delete(builder, "clone_from_vmxc_path")
			builder["clone_from_vmcx_path"] = path
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerHypervVmxcTypo) Synopsis() string {
	return `Fixes a typo replacing "clone_from_vmxc_path" with "clone_from_vmcx_path" ` +
		`in Hyper-V VMCX builder templates`
}
