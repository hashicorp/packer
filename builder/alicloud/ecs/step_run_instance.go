package ecs

import (
	"fmt"

	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepRunAlicloudInstance struct {
}

func (s *stepRunAlicloudInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)

	err := client.StartInstance(instance.InstanceId)
	if err != nil {
		err := fmt.Errorf("Error starting instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say("Starting instance.")
	err = client.WaitForInstance(instance.InstanceId, ecs.Running, ALICLOUD_DEFAULT_TIMEOUT)
	if err != nil {
		err := fmt.Errorf("Timeout waiting for instance to start: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRunAlicloudInstance) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if cancelled || halted {
		ui := state.Get("ui").(packer.Ui)
		client := state.Get("client").(*ecs.Client)
		instance := state.Get("instance").(*ecs.InstanceAttributesType)
		instanceAttrubite, _ := client.DescribeInstanceAttribute(instance.InstanceId)
		if instanceAttrubite.Status == ecs.Starting || instanceAttrubite.Status == ecs.Running {
			if err := client.StopInstance(instance.InstanceId, true); err != nil {
				ui.Say(fmt.Sprintf("Error stopping instance %s, it may still be around %s", instance.InstanceId, err))
				return
			}
			if err := client.WaitForInstance(instance.InstanceId, ecs.Stopped, ALICLOUD_DEFAULT_TIMEOUT); err != nil {
				ui.Say(fmt.Sprintf("Error stopping instance %s, it may still be around %s", instance.InstanceId, err))
			}
		}
	}
}
