package nutanix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	v3 "github.com/hashicorp/packer/builder/nutanix/common/v3"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepShutdownVM struct {
	Config *Config
}

func (s *stepShutdownVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmUUID := state.Get("vmUUID").(string)
	comm := state.Get("communicator").(packer.Communicator)

	ui.Say("Checking status of VM before powering down uuid: " + vmUUID)

	d := NewDriver(&s.Config.NutanixCluster, state)
	vmResponse, err := d.RetrieveReadyVM(ctx, 1*time.Minute)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	if *vmResponse.Spec.Resources.PowerState == "ON" {
		ui.Message("Issuing shutdown command.")
		d := NewDriver(&s.Config.NutanixCluster, state)

		if s.Config.Command != "" {
			ui.Say("Executing shutdown command...")
			log.Printf("Shutdown command: %s", s.Config.ShutdownConfig.Command)

			var stdout, stderr bytes.Buffer
			cmd := &packer.RemoteCmd{
				Command: s.Config.ShutdownConfig.Command,
				Stdout:  &stdout,
				Stderr:  &stderr,
			}
			err := comm.Start(ctx, cmd)
			if err != nil {
				state.Put("error", fmt.Errorf("Failed to send shutdown command: %s", err))
				return multistep.ActionHalt
			}

			shutdownTimer := time.After(s.Config.ShutdownConfig.Timeout)
			for {
				vmResponse, err := d.RetrieveReadyVM(ctx, 1*time.Minute)
				if err != nil {
					state.Put("error", err)
					return multistep.ActionHalt
				}
				if *vmResponse.Spec.Resources.PowerState != "ON" {
					break
				}

				select {
				case <-shutdownTimer:
					log.Printf("Shutdown stdout: %s", stdout.String())
					log.Printf("Shutdown stderr: %s", stderr.String())
					err := errors.New("Timeout while waiting for machine to shut down")
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				default:
					time.Sleep(1 * time.Second)
				}
			}
		} else {
			ui.Message("Shutdown command not provided, powering down from Nutanix...")
			vmRequest := &v3.VMIntentInput{
				Spec:     vmResponse.Spec,
				Metadata: vmResponse.Metadata,
			}
			*vmRequest.Spec.Resources.PowerState = "OFF"
			vmResponse, err = d.UpdateVM(ctx, vmRequest)
			if err != nil {
				state.Put("error", err)
				return multistep.ActionHalt
			}
		}

		log.Printf("Waiting max %s for shutdown to complete", s.Config.Timeout)

		ui.Message("VM is now powered off.")
	} else {
		ui.Message("Shutdown command not run, VM is already powered off.")
	}
	return multistep.ActionContinue
}

func (s *stepShutdownVM) Cleanup(state multistep.StateBag) {
}
