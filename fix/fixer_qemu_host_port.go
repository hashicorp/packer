// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerQEMUHostPort updates ssh_host_port_min and ssh_host_port_max to host_port_min and host_port_max for QEMU builders
type FixerQEMUHostPort struct{}

func (FixerQEMUHostPort) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		// replace ssh_host_port_min with host_port_min if it exists
		sshHostPortMin, ok := builder["ssh_host_port_min"]
		if ok {
			delete(builder, "ssh_host_port_min")
			builder["host_port_min"] = sshHostPortMin
		}

		// replace ssh_host_port_min with host_port_min if it exists
		sshHostPortMax, ok := builder["ssh_host_port_max"]
		if ok {
			delete(builder, "ssh_host_port_max")
			builder["host_port_max"] = sshHostPortMax
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerQEMUHostPort) Synopsis() string {
	return `Updates ssh_host_port_min and ssh_host_port_max to host_port_min and host_port_max`
}

func (FixerQEMUHostPort) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"transcend.qemu": []string{"ssh_host_port_max", "ssh_host_port_min"},
	}
}
