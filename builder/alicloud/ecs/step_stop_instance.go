package ecs

import (
	"context"
	"fmt"

	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepStopAlicloudInstance struct {
	ForceStop   bool
	DisableStop bool
}

func (s *stepStopAlicloudInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	ui := state.Get("ui").(packer.Ui)

	if !s.DisableStop {
		ui.Say(fmt.Sprintf("Stopping instance: %s", instance.InstanceId))
		err := client.StopInstance(instance.InstanceId, s.ForceStop)
		if err != nil {
			err := fmt.Errorf("Error stopping alicloud instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say(fmt.Sprintf("Waiting instance stopped: %s", instance.InstanceId))

	err := client.WaitForInstance(instance.InstanceId, ecs.Stopped, ALICLOUD_DEFAULT_TIMEOUT)
	if err != nil {
		err := fmt.Errorf("Error waiting for alicloud instance to stop: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStopAlicloudInstance) Cleanup(multistep.StateBag) {
	// No cleanup...
}
