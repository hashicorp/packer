// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

// FixerCommConfig removes ssh prefix from communicator port forwarding config
// for variables host_port_min, host_port_max, skip_nat_mapping
type FixerCommConfig struct{}

func (FixerCommConfig) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*": []string{"ssh_host_port_min", "ssh_host_port_max",
			"ssh_skip_nat_mapping"},
	}
}

func (FixerCommConfig) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	type template struct {
		Builders []interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.WeakDecode(input, &tpl); err != nil {
		return nil, err
	}

	for i, raw := range tpl.Builders {
		var builders map[string]interface{}
		if err := mapstructure.Decode(raw, &builders); err != nil {
			// Ignore errors, could be a non-map
			continue
		}

		// only virtualbox builders
		builderType := builders["type"].(string)
		if ok := strings.HasPrefix(builderType, "virtualbox"); !ok {
			continue
		}

		// ssh_host_port_min to host_port_min
		if _, ok := builders["host_port_min"]; ok {

			// drop ssh_host_port_min if it is also included
			if _, sshHostPortMinIncluded := builders["ssh_host_port_min"]; sshHostPortMinIncluded {
				delete(builders, "ssh_host_port_min")
			}

		} else if _, ok := builders["ssh_host_port_min"]; ok {

			// replace ssh_host_port_min with host_port_min
			sshHostPortMinRaw := builders["ssh_host_port_min"]
			delete(builders, "ssh_host_port_min")
			builders["host_port_min"] = sshHostPortMinRaw
		}

		// ssh_host_port_max to host_port_max
		if _, ok := builders["host_port_max"]; ok {

			// drop ssh_host_port_max if it is also included
			if _, sshHostPortMaxIncluded := builders["ssh_host_port_max"]; sshHostPortMaxIncluded {
				delete(builders, "ssh_host_port_max")
			}

		} else if _, ok := builders["ssh_host_port_max"]; ok {

			// replace ssh_host_port_max with host_port_max
			sshHostPortMaxRaw := builders["ssh_host_port_max"]
			delete(builders, "ssh_host_port_max")
			builders["host_port_max"] = sshHostPortMaxRaw

		}

		// ssh_skip_nat_mapping to skip_nat_mapping
		if _, ok := builders["skip_nat_mapping"]; ok {

			// drop ssh_skip_nat_mapping if it is also included
			if _, sshSkipNatMappingIncluded := builders["ssh_skip_nat_mapping"]; sshSkipNatMappingIncluded {
				delete(builders, "ssh_skip_nat_mapping")
			}

		} else if _, ok := builders["ssh_skip_nat_mapping"]; ok {

			// replace ssh_skip_nat_mapping with skip_nat_mapping
			sshSkipNatMappingRaw := builders["ssh_skip_nat_mapping"]
			sshSkipNatMappingBool, ok := sshSkipNatMappingRaw.(bool)
			if ok {
				delete(builders, "ssh_skip_nat_mapping")
				builders["skip_nat_mapping"] = sshSkipNatMappingBool
			}
		}

		// Write all changes back to template
		tpl.Builders[i] = builders
	}

	if len(tpl.Builders) > 0 {
		input["builders"] = tpl.Builders
	}

	return input, nil
}

func (FixerCommConfig) Synopsis() string {
	return `Remove ssh prefixes from communicator port forwarding configuration (host_port_min, host_port_max, skip_nat_mapping)`
}
