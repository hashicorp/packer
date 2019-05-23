package powershell

import (
	"testing"
)

func TestExecutionPolicy_Decode(t *testing.T) {
	config := map[string]interface{}{
		"inline":           []interface{}{"foo", "bar"},
		"execution_policy": "AllSigned",
	}
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatal(err)
	}

	if p.config.ExecutionPolicy != AllSigned {
		t.Fatalf("Expected AllSigned execution policy; got: %s", p.config.ExecutionPolicy)
	}
}
