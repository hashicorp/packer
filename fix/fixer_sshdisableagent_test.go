// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"reflect"
	"testing"
)

func TestFixerSSHDisableAgent_Impl(t *testing.T) {
	var _ Fixer = new(FixerSSHDisableAgent)
}

func TestFixerSSHDisableAgent_Fix(t *testing.T) {
	cases := []struct {
		Input    map[string]interface{}
		Expected map[string]interface{}
	}{
		// No disable_agent field
		{
			Input: map[string]interface{}{
				"type": "virtualbox",
			},

			Expected: map[string]interface{}{
				"type": "virtualbox",
			},
		},

		// disable_agent_forwarding without disable_agent
		{
			Input: map[string]interface{}{
				"ssh_disable_agent_forwarding": true,
			},

			Expected: map[string]interface{}{
				"ssh_disable_agent_forwarding": true,
			},
		},

		// disable_agent without disable_agent_forwarding
		{
			Input: map[string]interface{}{
				"ssh_disable_agent": true,
			},

			Expected: map[string]interface{}{
				"ssh_disable_agent_forwarding": true,
			},
		},

		// disable_agent and disable_agent_forwarding
		{
			Input: map[string]interface{}{
				"ssh_disable_agent":            true,
				"ssh_disable_agent_forwarding": false,
			},

			Expected: map[string]interface{}{
				"ssh_disable_agent_forwarding": false,
			},
		},
	}

	for _, tc := range cases {
		var f FixerSSHDisableAgent

		input := map[string]interface{}{
			"builders": []map[string]interface{}{tc.Input},
		}

		expected := map[string]interface{}{
			"builders": []map[string]interface{}{tc.Expected},
		}

		output, err := f.Fix(input)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if !reflect.DeepEqual(output, expected) {
			t.Fatalf("unexpected: %#v\nexpected: %#v\n", output, expected)
		}
	}
}
