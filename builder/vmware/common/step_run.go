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
	Headless           bool

	bootTime time.Time
	vmxPath  string
}

func (s *StepRun) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	// Set the VMX path so that we know we started the machine
	s.bootTime = time.Now()
	s.vmxPath = vmxPath

	ui.Say("Starting virtual machine...")
	if s.Headless {
		vncIpRaw, vncIpOk := state.GetOk("vnc_ip")
		vncPortRaw, vncPortOk := state.GetOk("vnc_port")

		if vncIpOk && vncPortOk {
			vncIp := vncIpRaw.(string)
			vncPort := vncPortRaw.(uint)

			ui.Message(fmt.Sprintf(
				"The VM will be run headless, without a GUI. If you want to\n"+
					"view the screen of the VM, connect via VNC without a password to\n"+
					"%s:%d", vncIp, vncPort))
		} else {
			ui.Message("The VM will be run headless, without a GUI, as configured.\n" +
				"If the run isn't succeeding as you expect, please enable the GUI\n" +
				"to inspect the progress of the build.")
		}
	}

	if err := driver.Start(vmxPath, s.Headless); err != nil {
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

	return multistep.ActionContinue
}

func (s *StepRun) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// If we started the machine... stop it.
	if s.vmxPath != "" {
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
		running, _ := driver.IsRunning(s.vmxPath)
		if running {
			ui.Say("Stopping virtual machine...")
			if err := driver.Stop(s.vmxPath); err != nil {
				ui.Error(fmt.Sprintf("Error stopping VM: %s", err))
			}
		}
	}
}
