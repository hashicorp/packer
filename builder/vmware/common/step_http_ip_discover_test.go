package common

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepHTTPIPDiscover_Run(t *testing.T) {
	state := testState(t)
	step := new(StepHTTPIPDiscover)
	driverMock := state.Get("driver").(Driver)
	hostIp, _ := driverMock.HostIP(state)
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

	// Halt step when fails to get ip
	state.Put("driver", &DriverMock{HostIPErr: errors.New("error")})
	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: step was supposed to fail %#v", action)
	}

	common.SetHTTPIP(previousHttpIp)
}
