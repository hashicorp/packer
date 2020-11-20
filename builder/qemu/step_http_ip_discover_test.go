package qemu

import (
	"bytes"
	"context"
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestStepHTTPIPDiscover_Run(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	config := &Config{}
	state.Put("config", config)
	step := new(stepHTTPIPDiscover)
	hostIp := "10.0.2.2"

	// Test the run
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
	httpIp := state.Get("http_ip").(string)
	if httpIp != hostIp {
		t.Fatalf("bad: Http ip is %s but was supposed to be %s", httpIp, hostIp)
	}
}
