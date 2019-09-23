package vminstance

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepStopVmInstance struct {
}

func (s *StepStopVmInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say(fmt.Sprintf("start stop zstack vminstance..."))

	uuid := getVmUuid(state)
	if uuid == "" {
		return halt(state, fmt.Errorf("cannot find vm instance %s", uuid), "")
	}

	if err := stopVmInstance(uuid, state); err != nil {
		return halt(state, err, "")
	}

	return multistep.ActionContinue
}

func stopVmInstance(uuid string, state multistep.StateBag) error {
	driver, config, _ := GetCommonFromState(state)

	err := driver.StopVmInstance(uuid)
	if err != nil {
		return err
	}

	errCh := driver.WaitForInstance("Stopped", uuid, "vm")
	select {
	case err = <-errCh:
		if err != nil {
			return err
		}
	case <-time.After(config.stateTimeout):
		return fmt.Errorf("time out after %v ms while waiting for vm instance to become Stopped", config.stateTimeout.Nanoseconds()/1000/1000)
	}

	return nil
}

func (s *StepStopVmInstance) Cleanup(state multistep.StateBag) {
	_, _, ui := GetCommonFromState(state)
	ui.Say("cleanup stop vm instance executing...")
}
