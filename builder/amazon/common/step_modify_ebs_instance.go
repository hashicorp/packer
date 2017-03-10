package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
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

	if s.EnableEnhancedNetworking {
		// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
		// As of February 2017, this applies to C3, C4, D2, I2, R3, and M4 (excluding m4.16xlarge)
		ui.Say("Enabling Enhanced Networking (SR-IOV)...")
		simple := "simple"
		_, err := ec2conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId:      instance.InstanceId,
			SriovNetSupport: &ec2.AttributeValue{Value: &simple},
		})
		if err != nil {
			err := fmt.Errorf("Error enabling Enhanced Networking (SR-IOV) on %s: %s", *instance.InstanceId, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set EnaSupport to true.
		// As of February 2017, this applies to C5, I3, P2, R4, X1, and m4.16xlarge
		ui.Say("Enabling Enhanced Networking (ENA)...")
		_, err = ec2conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId: instance.InstanceId,
			EnaSupport: &ec2.AttributeBooleanValue{Value: aws.Bool(true)},
		})
		if err != nil {
			err := fmt.Errorf("Error enabling Enhanced Networking (ENA) on %s: %s", *instance.InstanceId, err)
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
