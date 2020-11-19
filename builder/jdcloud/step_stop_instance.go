package jdcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/vm/apis"
)

type stepStopJDCloudInstance struct {
	InstanceSpecConfig *JDCloudInstanceSpecConfig
}

func (s *stepStopJDCloudInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Stopping this instance")

	req := apis.NewStopInstanceRequest(Region, s.InstanceSpecConfig.InstanceId)
	resp, err := VmClient.StopInstance(req)
	if err != nil || resp.Error.Code != FINE {
		ui.Error(fmt.Sprintf("[ERROR] Failed in trying to stop this vm: Error-%v ,Resp:%v", err, resp))
		return multistep.ActionHalt
	}

	_, err = InstanceStatusRefresher(s.InstanceSpecConfig.InstanceId, []string{VM_RUNNING, VM_STOPPING}, []string{VM_STOPPED})
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("Instance has been stopped :)")
	return multistep.ActionContinue
}

func (s *stepStopJDCloudInstance) Cleanup(multistep.StateBag) {
	return
}
