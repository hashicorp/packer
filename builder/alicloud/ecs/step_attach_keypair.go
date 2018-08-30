package ecs

import (
	"context"
	"fmt"

	"time"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepAttachKeyPair struct {
}

func (s *stepAttachKeyPair) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*ecs.Client)
	config := state.Get("config").(Config)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	timeoutPoint := time.Now().Add(120 * time.Second)
	keyPairName := config.Comm.SSHKeyPairName
	if keyPairName == "" {
		return multistep.ActionContinue
	}
	for {
		err := client.AttachKeyPair(&ecs.AttachKeyPairArgs{RegionId: common.Region(config.AlicloudRegion),
			KeyPairName: keyPairName, InstanceIds: "[\"" + instance.InstanceId + "\"]"})
		if err != nil {
			e, _ := err.(*common.Error)
			if (!(e.Code == "MissingParameter" || e.Code == "DependencyViolation.WindowsInstance" ||
				e.Code == "InvalidKeyPairName.NotFound" || e.Code == "InvalidRegionId.NotFound")) &&
				time.Now().Before(timeoutPoint) {
				time.Sleep(5 * time.Second)
				continue
			}
			err := fmt.Errorf("Error attaching keypair %s to instance %s : %s",
				keyPairName, instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		break
	}

	ui.Message(fmt.Sprintf("Attach keypair %s to instance: %s", keyPairName, instance.InstanceId))

	return multistep.ActionContinue
}

func (s *stepAttachKeyPair) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*ecs.Client)
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	keyPairName := config.Comm.SSHKeyPairName
	if keyPairName == "" {
		return
	}

	err := client.DetachKeyPair(&ecs.DetachKeyPairArgs{RegionId: common.Region(config.AlicloudRegion),
		KeyPairName: keyPairName, InstanceIds: "[\"" + instance.InstanceId + "\"]"})
	if err != nil {
		err := fmt.Errorf("Error Detaching keypair %s to instance %s : %s", keyPairName,
			instance.InstanceId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return
	}

	ui.Message(fmt.Sprintf("Detach keypair %s from instance: %s", keyPairName, instance.InstanceId))

}
