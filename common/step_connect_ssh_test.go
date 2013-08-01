package common

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepConnectSSH_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepConnectSSH)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("connect ssh should be a step")
	}
}
