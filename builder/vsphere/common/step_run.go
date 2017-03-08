package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step runs the created virtual machine.
//
// Uses:
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepRun struct {
	BootWait           time.Duration
	DurationBeforeStop time.Duration

	bootTime time.Time
	started  bool
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	s.started = false
	s.bootTime = time.Now()

	ui.Say("Starting virtual machine...")

	if err := driver.Start(); err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait the wait amount
	if int64(s.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", s.BootWait.String()))
		wait := time.After(s.BootWait)
	WAITLOOP:
		for {
			select {
			case <-wait:
				break WAITLOOP
			case <-time.After(1 * time.Second):
				if _, ok := state.GetOk(multistep.StateCancelled); ok {
					return multistep.ActionHalt
				}
			}
		}

	}
	s.started = true

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// If we started the machine... stop it.
	if s.started {
		// If we started it less than 5 seconds ago... wait.
		sinceBootTime := time.Since(s.bootTime)
		waitBootTime := s.DurationBeforeStop
		if sinceBootTime < waitBootTime {
			sleepTime := waitBootTime - sinceBootTime
			ui.Say(fmt.Sprintf(
				"Waiting %s to give VMware time to clean up...", sleepTime.String()))
			time.Sleep(sleepTime)
		}

		// See if it is running
		running, _ := driver.IsRunning()
		if running {
			ui.Say("Stopping virtual machine...")
			if err := driver.Stop(); err != nil {
				ui.Error(fmt.Sprintf("Error stopping VM: %s", err))
			}
		}
	}
}
