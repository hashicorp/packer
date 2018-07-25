package common

import (
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepProvision_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepProvision)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("provision should be a step")
	}
}
