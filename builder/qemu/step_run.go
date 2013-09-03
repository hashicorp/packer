package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"strings"
	"time"
)

type stepRun struct {
	vmName string
}

func runBootCommand(state multistep.StateBag,
	actionChannel chan multistep.StepAction) {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)
	bootCmd := stepTypeBootCommand{}

	if int64(config.bootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.bootWait))
		time.Sleep(config.bootWait)
	}

	actionChannel <- bootCmd.Run(state)
}

func cancelCallback(state multistep.StateBag) bool {
	cancel := false
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		cancel = true
	}
	return cancel
}

func (s *stepRun) runVM(
	sendBootCommands bool,
	bootDrive string,
	state multistep.StateBag) multistep.StepAction {

	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := config.VMName

	imgPath := filepath.Join(config.OutputDir,
		fmt.Sprintf("%s.%s", vmName, strings.ToLower(config.Format)))
	isoPath := state.Get("iso_path").(string)
	vncPort := state.Get("vnc_port").(uint)
	guiArgument := "sdl"
	sshHostPort := state.Get("sshHostPort").(uint)
	vnc := fmt.Sprintf("0.0.0.0:%d", vncPort-5900)

	ui.Say("Starting the virtual machine for OS Install...")
	if config.Headless == true {
		ui.Message("WARNING: The VM will be started in headless mode, as configured.\n" +
			"In headless mode, errors during the boot sequence or OS setup\n" +
			"won't be easily visible. Use at your own discretion.")
		guiArgument = "none"
	}

	command := []string{
		"-name", vmName,
		"-machine", fmt.Sprintf("type=pc-1.0,accel=%s", config.Accelerator),
		"-display", guiArgument,
		"-net", fmt.Sprintf("nic,model=%s", config.NetDevice),
		"-net", "user",
		"-drive", fmt.Sprintf("file=%s,if=%s", imgPath, config.DiskInterface),
		"-cdrom", isoPath,
		"-boot", bootDrive,
		"-m", "512m",
		"-redir", fmt.Sprintf("tcp:%v::22", sshHostPort),
		"-vnc", vnc,
	}
	if err := driver.Qemu(vmName, command...); err != nil {
		err := fmt.Errorf("Error launching VM: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	// run the boot command after its own timeout
	if sendBootCommands {
		waitDone := make(chan multistep.StepAction, 1)
		go runBootCommand(state, waitDone)
		select {
		case action := <-waitDone:
			if action != multistep.ActionContinue {
				// stop the VM in its tracks
				driver.Stop(vmName)
				return multistep.ActionHalt
			}
		}
	}

	ui.Say("Waiting for VM to shutdown...")
	if err := driver.WaitForShutdown(vmName, sendBootCommands, state, cancelCallback); err != nil {
		err := fmt.Errorf("Error waiting for initial VM install to shutdown: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRun) Run(state multistep.StateBag) multistep.StepAction {
	// First, the OS install boot
	action := s.runVM(true, "d", state)

	if action == multistep.ActionContinue {
		// Then the provisioning install
		action = s.runVM(false, "c", state)
	}

	return action
}

func (s *stepRun) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.Stop(s.vmName); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}
