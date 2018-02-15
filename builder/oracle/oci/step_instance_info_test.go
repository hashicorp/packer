package oci

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestInstanceInfo(t *testing.T) {
	state := testState()
	state.Put("instance_id", "ocid1...")

	step := new(stepInstanceInfo)
	defer step.Cleanup(state)

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	instanceIPRaw, ok := state.GetOk("instance_ip")
	if !ok {
		t.Fatalf("should have instance_ip")
	}

	if instanceIPRaw.(string) != "ip" {
		t.Fatalf("should've got ip ('%s' != 'ip')", instanceIPRaw.(string))
	}
}

func TestInstanceInfo_GetInstanceIPErr(t *testing.T) {
	state := testState()
	state.Put("instance_id", "ocid1...")

	step := new(stepInstanceInfo)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)
	driver.GetInstanceIPErr = errors.New("error")

	if action := step.Run(context.Background(), state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("instance_ip"); ok {
		t.Fatalf("should NOT have instance_ip")
	}
}
