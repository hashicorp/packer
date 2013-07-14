package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step runs the created virtual machine.
//
// Uses:
//   config *config
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type stepRun struct {
	bootTime time.Time
	vmxPath  string
}

func (s *stepRun) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)
	vncPort := state["vnc_port"].(uint)

	// Set the VMX path so that we know we started the machine
	s.bootTime = time.Now()
	s.vmxPath = vmxPath

	ui.Say("Starting virtual machine...")
	if config.Headless {
		ui.Message(fmt.Sprintf(
			"The VM will be run headless, without a GUI. If you want to\n"+
				"view the screen of the VM, connect via VNC without a password to\n"+
				"127.0.0.1:%d", vncPort))
	}

	if err := driver.Start(vmxPath, config.Headless); err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait the wait amount
	if int64(config.bootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.bootWait.String()))
		time.Sleep(config.bootWait)
	}

	return multistep.ActionContinue
}

func (s *stepRun) Cleanup(state map[string]interface{}) {
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	// If we started the machine... stop it.
	if s.vmxPath != "" {
		// If we started it less than 5 seconds ago... wait.
		sinceBootTime := time.Since(s.bootTime)
		waitBootTime := 5 * time.Second
		if sinceBootTime < waitBootTime {
			sleepTime := waitBootTime - sinceBootTime
			ui.Say(fmt.Sprintf("Waiting %s to give VMware time to clean up...", sleepTime.String()))
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
