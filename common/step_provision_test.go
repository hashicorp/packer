package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepProvision_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepProvision)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("provision should be a step")
	}
}
