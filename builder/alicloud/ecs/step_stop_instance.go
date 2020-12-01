package ecs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepStopAlicloudInstance struct {
	ForceStop   bool
	DisableStop bool
}

func (s *stepStopAlicloudInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	instance := state.Get("instance").(*ecs.Instance)
	ui := state.Get("ui").(packersdk.Ui)

	if !s.DisableStop {
		ui.Say(fmt.Sprintf("Stopping instance: %s", instance.InstanceId))

		stopInstanceRequest := ecs.CreateStopInstanceRequest()
		stopInstanceRequest.InstanceId = instance.InstanceId
		stopInstanceRequest.ForceStop = requests.Boolean(strconv.FormatBool(s.ForceStop))
		if _, err := client.StopInstance(stopInstanceRequest); err != nil {
			return halt(state, err, "Error stopping alicloud instance")
		}
	}

	ui.Say(fmt.Sprintf("Waiting instance stopped: %s", instance.InstanceId))

	_, err := client.WaitForInstanceStatus(instance.RegionId, instance.InstanceId, InstanceStatusStopped)
	if err != nil {
		return halt(state, err, "Error waiting for alicloud instance to stop")
	}

	return multistep.ActionContinue
}

func (s *stepStopAlicloudInstance) Cleanup(multistep.StateBag) {
	// No cleanup...
}
