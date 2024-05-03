// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package powershell

import (
	"testing"
)

func TestExecutionPolicy_Decode(t *testing.T) {
	config := map[string]interface{}{
		"inline":           []interface{}{"foo", "bar"},
		"execution_policy": "allsigned",
	}
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatal(err)
	}

	if p.config.ExecutionPolicy != ExecutionPolicyAllsigned {
		t.Fatalf("Expected AllSigned execution policy; got: %s", p.config.ExecutionPolicy)
	}
}
