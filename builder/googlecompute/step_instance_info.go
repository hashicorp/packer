package googlecompute

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// stepInstanceInfo represents a Packer build step that gathers GCE instance info.
type StepInstanceInfo struct {
	Debug bool

	info int
}

// Run executes the Packer build step that gathers GCE instance info.
// This adds "instance_ip" to the multistep state.
func (s *StepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	instanceName := state.Get("instance_name").(string)

	ui.Say("Waiting for the instance to become running...")
	errCh := driver.WaitForInstance("RUNNING", config.Zone, instanceName)
	var err error
	select {
	case err = <-errCh:
	case <-time.After(config.stateTimeout):
		err = errors.New("time out while waiting for instance to become running")
	}

	if err != nil {
		err := fmt.Errorf("Error waiting for instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if config.UseInternalIP {
		ip, err := driver.GetInternalIP(config.Zone, instanceName)
		if err != nil {
			err := fmt.Errorf("Error retrieving instance internal ip address: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if s.Debug {
			if ip != "" {
				ui.Message(fmt.Sprintf("Internal IP: %s", ip))
			}
		}
		ui.Message(fmt.Sprintf("IP: %s", ip))
		state.Put("instance_ip", ip)
		return multistep.ActionContinue
	} else {
		ip, err := driver.GetNatIP(config.Zone, instanceName)
		if err != nil {
			err := fmt.Errorf("Error retrieving instance nat ip address: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if s.Debug {
			if ip != "" {
				ui.Message(fmt.Sprintf("Public IP: %s", ip))
			}
		}
		ui.Message(fmt.Sprintf("IP: %s", ip))
		state.Put("instance_ip", ip)
		return multistep.ActionContinue
	}
}

// Cleanup.
func (s *StepInstanceInfo) Cleanup(state multistep.StateBag) {}
