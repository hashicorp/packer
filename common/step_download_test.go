package common

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepDownload_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepDownload)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("download should be a step")
	}
}
