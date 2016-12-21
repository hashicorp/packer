package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepModifyEBSBackedInstance struct {
	EnhancedNetworkingType string
}

func (s *StepModifyEBSBackedInstance) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	// Set EnaSupport to true.
	if s.EnhancedNetworkingType == "ena" {
		ui.Say("Enabling ENA support...")
		truthy := true
		_, err := ec2conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId: instance.InstanceId,
			EnaSupport: &ec2.AttributeBooleanValue{Value: &truthy},
		})
		if err != nil {
			err := fmt.Errorf("Error enabling ENA support on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
	if s.EnhancedNetworkingType == "sriov" {
		ui.Say("Enabling SR-IOV support...")
		simple := "simple"
		_, err := ec2conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId:      instance.InstanceId,
			SriovNetSupport: &ec2.AttributeValue{Value: &simple},
		})
		if err != nil {
			err := fmt.Errorf("Error enabling SR-IOV support on %s: %s", *instance.InstanceId, err)
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
