package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
	"io/ioutil"
)

// This step starts the virtual machine.
//
// Uses:
//
// Produces:
type stepRun struct {
	vmName string
}

func (s *stepRun) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Starting the virtual machine...")
	guiArgument := "gui"
	if config.Headless == true {
		ui.Message("WARNING: The VM will be started in headless mode, as configured.\n" +
			"In headless mode, errors during the boot sequence or OS setup\n" +
			"won't be easily visible. Use at your own discretion.")
		guiArgument = "headless"
	}
	command := []string{"startvm", vmName, "--type", guiArgument}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmName = vmName

	if int64(config.bootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.bootWait))
		time.Sleep(config.bootWait)
	}
	// read the vbox file, and store it as a string
	dir := fmt.Sprintf("\"$HOME/VirtualBox VMs/packer-Vbox test/%s.vbox\"", s.vmName)
	b, err := ioutil.ReadFile(dir)
	tempString := string(b)
	if err != nil {
		err := fmt.Errorf("Error reading file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// descript is what we want to add to the vbox file
	descript := fmt.Sprintf("<Description>%s</Description>", config.Description)
	// Find the fourth occurrence of ">" to get the correct location to insert "descript"
	count := 0
	newString := ""
	for _, sub := range tempString {
		newString += string(sub)
		if string(sub) == ">" {
			count++
		}
		if count == 4 {
			newString += descript
			count = 5
		}
	}
	err = ioutil.WriteFile(dir, []byte(newString), 0644)
	if err != nil {
		err := fmt.Errorf("Error writing to file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRun) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if running, _ := driver.IsRunning(s.vmName); running {
		if err := driver.VBoxManage("controlvm", s.vmName, "poweroff"); err != nil {
			ui.Error(fmt.Sprintf("Error shutting down VM: %s", err))
		}
	}
}
