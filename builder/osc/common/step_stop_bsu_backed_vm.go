package common

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/packer/builder/osc/common/retry"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

type StepStopBSUBackedVm struct {
	Skip          bool
	DisableStopVm bool
}

func (s *StepStopBSUBackedVm) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	oscconn := state.Get("osc").(*osc.APIClient)
	vm := state.Get("vm").(osc.Vm)
	ui := state.Get("ui").(packersdk.Ui)

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
		err := retry.Run(10, 60, 6, func(i uint) (bool, error) {
			ui.Message(fmt.Sprintf("Stopping vm, attempt %d", i+1))

			_, _, err = oscconn.VmApi.StopVms(context.Background(), &osc.StopVmsOpts{
				StopVmsRequest: optional.NewInterface(osc.StopVmsRequest{
					VmIds: []string{vm.VmId},
				}),
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
	err = waitUntilOscVmStopped(oscconn, vm.VmId)

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
