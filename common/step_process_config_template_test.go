package common

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepProcessConfigTemplate_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepProcessConfigTemplate)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("StepProcessConfigTemplate should be a step")
	}
}
