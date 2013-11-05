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

func (s *stepRun) getCommandArgs(
	bootDrive string,
	state multistep.StateBag) []string {

	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*config)
	vmName := config.VMName
	imgPath := filepath.Join(config.OutputDir,
		fmt.Sprintf("%s.%s", vmName, strings.ToLower(config.Format)))
	isoPath := state.Get("iso_path").(string)
	vncPort := state.Get("vnc_port").(uint)
	guiArgument := "sdl"
	sshHostPort := state.Get("sshHostPort").(uint)
	vnc := fmt.Sprintf("0.0.0.0:%d", vncPort-5900)

	if config.Headless == true {
		ui.Message("WARNING: The VM will be started in headless mode, as configured.\n" +
			"In headless mode, errors during the boot sequence or OS setup\n" +
			"won't be easily visible. Use at your own discretion.")
		guiArgument = "none"
	}

	defaultArgs := make(map[string]string)
	defaultArgs["-name"] = vmName
	defaultArgs["-machine"] = fmt.Sprintf("type=pc-1.0,accel=%s", config.Accelerator)
	defaultArgs["-display"] = guiArgument
	defaultArgs["-netdev"] = "user,id=user.0"
	defaultArgs["-device"] = fmt.Sprintf("%s,netdev=user.0", config.NetDevice)
	defaultArgs["-drive"] = fmt.Sprintf("file=%s,if=%s", imgPath, config.DiskInterface)
	defaultArgs["-cdrom"] = isoPath
	defaultArgs["-boot"] = bootDrive
	defaultArgs["-m"] = "512m"
	defaultArgs["-redir"] = fmt.Sprintf("tcp:%v::22", sshHostPort)
	defaultArgs["-vnc"] = vnc

	inArgs := make(map[string][]string)
	if len(config.QemuArgs) > 0 {
		ui.Say("Overriding defaults Qemu arguments with QemuArgs...")

		// becuase qemu supports multiple appearances of the same
		// switch, just different values, each key in the args hash
		// will have an array of string values
		for _, qemuArgs := range config.QemuArgs {
			key := qemuArgs[0]
			val := strings.Join(qemuArgs[1:], "")
			if _, ok := inArgs[key]; !ok {
				inArgs[key] = make([]string, 0)
			}
			if len(val) > 0 {
				inArgs[key] = append(inArgs[key], val)
			}
		}
	}

	// get any remaining missing default args from the default settings
	for key := range defaultArgs {
		if _, ok := inArgs[key]; !ok {
			arg := make([]string, 1)
			arg[0] = defaultArgs[key]
			inArgs[key] = arg
		}
	}

	// Flatten to array of strings
	outArgs := make([]string, 0)
	for key, values := range inArgs {
		if len(values) > 0 {
			for idx := range values {
				outArgs = append(outArgs, key, values[idx])
			}
		} else {
			outArgs = append(outArgs, key)
		}
	}

	return outArgs
}

func (s *stepRun) runVM(
	sendBootCommands bool,
	bootDrive string,
	state multistep.StateBag) multistep.StepAction {

	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := config.VMName

	ui.Say("Starting the virtual machine for OS Install...")
	command := s.getCommandArgs(bootDrive, state)
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
