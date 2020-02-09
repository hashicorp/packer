package common

import (
	"context"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"testing"
)

func TestStepHTTPIPDiscover_Run(t *testing.T) {
	state := new(multistep.BasicStateBag)
	step := new(StepHTTPIPDiscover)
	hostIp := "10.0.2.2"
	previousHttpIp := common.GetHTTPIP()

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	httpIp := common.GetHTTPIP()
	if httpIp != hostIp {
		t.Fatalf("bad: Http ip is %s but was supposed to be %s", httpIp, hostIp)
	}

	common.SetHTTPIP(previousHttpIp)
}
