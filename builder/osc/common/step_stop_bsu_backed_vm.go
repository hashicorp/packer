package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

type StepStopBSUBackedVm struct {
	Skip          bool
	DisableStopVm bool
}

func (s *StepStopBSUBackedVm) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	vm := state.Get("vm").(oapi.Vm)
	ui := state.Get("ui").(packer.Ui)

	// Skip when it is a spot vm
	if s.Skip {
		return multistep.ActionContinue
	}

	var err error

	if !s.DisableStopVm {
		// Stop the vm so we can create an AMI from it
		ui.Say("Stopping the source vm...")

		// Amazon EC2 API follows an eventual consistency model.

		// This means that if you run a command to modify or describe a resource
		// that you just created, its ID might not have propagated throughout
		// the system, and you will get an error responding that the resource
		// does not exist.

		// Work around this by retrying a few times, up to about 5 minutes.
		err := common.Retry(10, 60, 6, func(i uint) (bool, error) {
			ui.Message(fmt.Sprintf("Stopping vm, attempt %d", i+1))

			_, err = oapiconn.POST_StopVms(oapi.StopVmsRequest{
				VmIds: []string{vm.VmId},
			})

			if err == nil {
				// success
				return true, nil
			}

			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidVmID.NotFound" {
					ui.Message(fmt.Sprintf(
						"Error stopping vm; will retry ..."+
							"Error: %s", err))
					// retry
					return false, nil
				}
			}
			// errored, but not in expected way. Don't want to retry
			return true, err
		})

		if err != nil {
			err := fmt.Errorf("Error stopping vm: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	} else {
		ui.Say("Automatic vm stop disabled. Please stop vm manually.")
	}

	// Wait for the vm to actually stop
	ui.Say("Waiting for the vm to stop...")
	err = waitUntilVmStopped(oapiconn, vm.VmId)

	if err != nil {
		err := fmt.Errorf("Error waiting for vm to stop: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepStopBSUBackedVm) Cleanup(multistep.StateBag) {
	// No cleanup...
}
