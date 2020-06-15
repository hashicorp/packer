package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepHTTPIPDiscover_Run(t *testing.T) {
	state := new(multistep.BasicStateBag)
	step := new(StepHTTPIPDiscover)

	// without setting HTTPIP
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	_, ok := state.GetOk("http_ip")
	if !ok {
		t.Fatal("should have http_ip")
	}

	// setting HTTPIP
	ip := "10.0.2.2"
	step = &StepHTTPIPDiscover{
		HTTPIP: ip,
	}
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	httpIp, ok := state.GetOk("http_ip")
	if !ok {
		t.Fatal("should have http_ip")
	}
	if httpIp != ip {
		t.Fatalf("bad: Http ip is %s but was supposed to be %s", httpIp, ip)
	}
}
