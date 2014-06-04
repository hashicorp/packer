package ebs

import (
	"fmt"

	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepModifyInstance struct{}

func (s *stepModifyInstance) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
	if config.AMIEnhancedNetworking {
		ui.Say("Enabling Enhanced Networking...")
		_, err := ec2conn.ModifyInstance(
			instance.InstanceId,
			&ec2.ModifyInstance{SriovNetSupport: true},
		)
		if err != nil {
			err := fmt.Errorf("Error enabling Enhanced Networking on %s: %s", instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepModifyInstance) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
