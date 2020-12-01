package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepAttachKeyPair struct {
}

var attachKeyPairNotRetryErrors = []string{
	"MissingParameter",
	"DependencyViolation.WindowsInstance",
	"InvalidKeyPairName.NotFound",
	"InvalidRegionId.NotFound",
}

func (s *stepAttachKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	instance := state.Get("instance").(*ecs.Instance)
	keyPairName := config.Comm.SSHKeyPairName
	if keyPairName == "" {
		return multistep.ActionContinue
	}

	_, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateAttachKeyPairRequest()
			request.RegionId = config.AlicloudRegion
			request.KeyPairName = keyPairName
			request.InstanceIds = "[\"" + instance.InstanceId + "\"]"
			return client.AttachKeyPair(request)
		},
		EvalFunc: client.EvalCouldRetryResponse(attachKeyPairNotRetryErrors, EvalNotRetryErrorType),
	})

	if err != nil {
		return halt(state, err, fmt.Sprintf("Error attaching keypair %s to instance %s", keyPairName, instance.InstanceId))
	}

	ui.Message(fmt.Sprintf("Attach keypair %s to instance: %s", keyPairName, instance.InstanceId))
	return multistep.ActionContinue
}

func (s *stepAttachKeyPair) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*ClientWrapper)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	instance := state.Get("instance").(*ecs.Instance)
	keyPairName := config.Comm.SSHKeyPairName
	if keyPairName == "" {
		return
	}

	detachKeyPairRequest := ecs.CreateDetachKeyPairRequest()
	detachKeyPairRequest.RegionId = config.AlicloudRegion
	detachKeyPairRequest.KeyPairName = keyPairName
	detachKeyPairRequest.InstanceIds = fmt.Sprintf("[\"%s\"]", instance.InstanceId)
	_, err := client.DetachKeyPair(detachKeyPairRequest)
	if err != nil {
		err := fmt.Errorf("Error Detaching keypair %s to instance %s : %s", keyPairName,
			instance.InstanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return
	}

	ui.Message(fmt.Sprintf("Detach keypair %s from instance: %s", keyPairName, instance.InstanceId))

}
