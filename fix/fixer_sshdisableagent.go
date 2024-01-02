// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerSSHDisableAgent changes the "ssh_disable_agent" of a template
// to "ssh_disable_agent_forwarding".
type FixerSSHDisableAgent struct{}

func (FixerSSHDisableAgent) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*": []string{"ssh_disable_agent"},
	}
}

func (FixerSSHDisableAgent) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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
		sshDisableAgentRaw, ok := builder["ssh_disable_agent"]
		if !ok {
			continue
		}

		sshDisableAgent, ok := sshDisableAgentRaw.(bool)
		if !ok {
			continue
		}

		// only assign to ssh_disable_agent_forwarding if it doesn't
		// already exist; otherwise we'll just ignore ssh_disable_agent
		_, sshDisableAgentIncluded := builder["ssh_disable_agent_forwarding"]
		if !sshDisableAgentIncluded {
			builder["ssh_disable_agent_forwarding"] = sshDisableAgent
		}

		delete(builder, "ssh_disable_agent")
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerSSHDisableAgent) Synopsis() string {
	return `Updates builders using "ssh_disable_agent" to use "ssh_disable_agent_forwarding"`
}
