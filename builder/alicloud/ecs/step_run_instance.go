package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepRunAlicloudInstance struct {
}

func (s *stepRunAlicloudInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)
	instance := state.Get("instance").(*ecs.Instance)

	startInstanceRequest := ecs.CreateStartInstanceRequest()
	startInstanceRequest.InstanceId = instance.InstanceId
	if _, err := client.StartInstance(startInstanceRequest); err != nil {
		return halt(state, err, "Error starting instance")
	}

	ui.Say(fmt.Sprintf("Starting instance: %s", instance.InstanceId))

	_, err := client.WaitForInstanceStatus(instance.RegionId, instance.InstanceId, InstanceStatusRunning)
	if err != nil {
		return halt(state, err, "Timeout waiting for instance to start")
	}

	return multistep.ActionContinue
}

func (s *stepRunAlicloudInstance) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*ClientWrapper)
	instance := state.Get("instance").(*ecs.Instance)

	describeInstancesRequest := ecs.CreateDescribeInstancesRequest()
	describeInstancesRequest.InstanceIds = fmt.Sprintf("[\"%s\"]", instance.InstanceId)
	instancesResponse, _ := client.DescribeInstances(describeInstancesRequest)

	if len(instancesResponse.Instances.Instance) == 0 {
		return
	}

	instanceAttribute := instancesResponse.Instances.Instance[0]
	if instanceAttribute.Status == InstanceStatusStarting || instanceAttribute.Status == InstanceStatusRunning {
		stopInstanceRequest := ecs.CreateStopInstanceRequest()
		stopInstanceRequest.InstanceId = instance.InstanceId
		stopInstanceRequest.ForceStop = requests.NewBoolean(true)
		if _, err := client.StopInstance(stopInstanceRequest); err != nil {
			ui.Say(fmt.Sprintf("Error stopping instance %s, it may still be around %s", instance.InstanceId, err))
			return
		}

		_, err := client.WaitForInstanceStatus(instance.RegionId, instance.InstanceId, InstanceStatusStopped)
		if err != nil {
			ui.Say(fmt.Sprintf("Error stopping instance %s, it may still be around %s", instance.InstanceId, err))
		}
	}
}
