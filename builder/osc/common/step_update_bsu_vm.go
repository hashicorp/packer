package common

import (
	"context"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

type StepUpdateBSUBackedVm struct {
	EnableAMIENASupport      *bool
	EnableAMISriovNetSupport bool
}

func (s *StepUpdateBSUBackedVm) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// oapiconn := state.Get("oapi").(*oapi.Client)
	// vm := state.Get("vm").(*oapi.Vm)
	// ui := state.Get("ui").(packersdk.Ui)

	// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
	// As of February 2017, this applies to C3, C4, D2, I2, R3, and M4 (excluding m4.16xlarge)
	// if s.EnableAMISriovNetSupport {
	// 	ui.Say("Enabling Enhanced Networking (SR-IOV)...")
	// 	simple := "simple"
	// 	_, err := oapiconn.POST_UpdateVm(oapi.UpdateVmRequest{
	// 		VmId:            vm.VmId,
	// 		SriovNetSupport: &oapi.AttributeValue{Value: &simple},
	// 	})
	// 	if err != nil {
	// 		err := fmt.Errorf("Error enabling Enhanced Networking (SR-IOV) on %s: %s", *vm.VmId, err)
	// 		state.Put("error", err)
	// 		ui.Error(err.Error())
	// 		return multistep.ActionHalt
	// 	}
	// }

	// Handle EnaSupport flag.
	// As of February 2017, this applies to C5, I3, P2, R4, X1, and m4.16xlarge
	// if s.EnableAMIENASupport != nil {
	// 	var prefix string
	// 	if *s.EnableAMIENASupport {
	// 		prefix = "En"
	// 	} else {
	// 		prefix = "Dis"
	// 	}
	// 	ui.Say(fmt.Sprintf("%sabling Enhanced Networking (ENA)...", prefix))
	// 	_, err := oapiconn.UpdateVmAttribute(&oapi.UpdateVmAttributeInput{
	// 		VmId:       vm.VmId,
	// 		EnaSupport: &oapi.AttributeBooleanValue{Value: aws.Bool(*s.EnableAMIENASupport)},
	// 	})
	// 	if err != nil {
	// 		err := fmt.Errorf("Error %sabling Enhanced Networking (ENA) on %s: %s", strings.ToLower(prefix), *vm.VmId, err)
	// 		state.Put("error", err)
	// 		ui.Error(err.Error())
	// 		return multistep.ActionHalt
	// 	}
	// }

	return multistep.ActionContinue
}

func (s *StepUpdateBSUBackedVm) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
