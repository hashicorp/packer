// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerSSHTimout replaces ssh_wait_timeout with ssh_timeout
type FixerSSHTimout struct{}

func (FixerSSHTimout) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*": []string{"ssh_wait_timeout"},
	}
}

func (FixerSSHTimout) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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

		if _, ok := builders["ssh_timeout"]; ok {

			// drop ssh_wait_timeout if it is also included
			if _, sshWaitTimeoutIncluded := builders["ssh_wait_timeout"]; sshWaitTimeoutIncluded {
				delete(builders, "ssh_wait_timeout")
			}

		} else {

			// replace ssh_wait_timeout with ssh_timeout if it exists
			sshWaitTimeoutRaw, ok := builders["ssh_wait_timeout"]
			if !ok {
				continue
			}

			sshWaitTimeoutString, ok := sshWaitTimeoutRaw.(string)
			if !ok {
				continue
			}

			delete(builders, "ssh_wait_timeout")
			builders["ssh_timeout"] = sshWaitTimeoutString
		}

		// Write all changes back to template
		tpl.Builders[i] = builders
	}

	if len(tpl.Builders) > 0 {
		input["builders"] = tpl.Builders
	}

	return input, nil
}

func (FixerSSHTimout) Synopsis() string {
	return `Replaces "ssh_wait_timeout" with "ssh_timeout"`
}
