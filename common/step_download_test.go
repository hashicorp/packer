package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepDownload_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepDownload)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("download should be a step")
	}
}
