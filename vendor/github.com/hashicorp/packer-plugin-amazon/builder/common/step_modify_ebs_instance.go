package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	confighelper "github.com/hashicorp/packer-plugin-sdk/template/config"
)

type StepModifyEBSBackedInstance struct {
	Skip                     bool
	EnableAMIENASupport      confighelper.Trilean
	EnableAMISriovNetSupport bool
}

func (s *StepModifyEBSBackedInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(ec2iface.EC2API)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packersdk.Ui)

	// Skip when it is a spot instance
	if s.Skip {
		return multistep.ActionContinue
	}

	// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
	// As of February 2017, this applies to C3, C4, D2, I2, R3, and M4 (excluding m4.16xlarge)
	if s.EnableAMISriovNetSupport {
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
	}

	// Handle EnaSupport flag.
	// As of February 2017, this applies to C5, I3, P2, R4, X1, and m4.16xlarge
	if s.EnableAMIENASupport != confighelper.TriUnset {
		var prefix string
		if s.EnableAMIENASupport.True() {
			prefix = "En"
		} else {
			prefix = "Dis"
		}
		ui.Say(fmt.Sprintf("%sabling Enhanced Networking (ENA)...", prefix))
		_, err := ec2conn.ModifyInstanceAttribute(&ec2.ModifyInstanceAttributeInput{
			InstanceId: instance.InstanceId,
			EnaSupport: &ec2.AttributeBooleanValue{Value: s.EnableAMIENASupport.ToBoolPointer()},
		})
		if err != nil {
			err := fmt.Errorf("Error %sabling Enhanced Networking (ENA) on %s: %s", strings.ToLower(prefix), *instance.InstanceId, err)
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
