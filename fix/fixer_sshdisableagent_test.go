// Copyright IBM Corp. 2024, 2025
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
		Input    map[string]any
		Expected map[string]any
	}{
		// No disable_agent field
		{
			Input: map[string]any{
				"type": "virtualbox",
			},

			Expected: map[string]any{
				"type": "virtualbox",
			},
		},

		// disable_agent_forwarding without disable_agent
		{
			Input: map[string]any{
				"ssh_disable_agent_forwarding": true,
			},

			Expected: map[string]any{
				"ssh_disable_agent_forwarding": true,
			},
		},

		// disable_agent without disable_agent_forwarding
		{
			Input: map[string]any{
				"ssh_disable_agent": true,
			},

			Expected: map[string]any{
				"ssh_disable_agent_forwarding": true,
			},
		},

		// disable_agent and disable_agent_forwarding
		{
			Input: map[string]any{
				"ssh_disable_agent":            true,
				"ssh_disable_agent_forwarding": false,
			},

			Expected: map[string]any{
				"ssh_disable_agent_forwarding": false,
			},
		},
	}

	for _, tc := range cases {
		var f FixerSSHDisableAgent

		input := map[string]any{
			"builders": []map[string]any{tc.Input},
		}

		expected := map[string]any{
			"builders": []map[string]any{tc.Expected},
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
