package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepModifyEBSBackedInstance struct {
	EnableEnhancedNetworking bool
}

func (s *StepModifyEBSBackedInstance) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
	if s.EnableEnhancedNetworking {
		ui.Say("Enabling Enhanced Networking...")
		simple := "simple"
		_, err := ec2conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId:      instance.InstanceId,
			SriovNetSupport: &ec2.AttributeValue{Value: &simple},
		})
		if err != nil {
			err := fmt.Errorf("Error enabling Enhanced Networking on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepModifyEBSBackedInstance) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
