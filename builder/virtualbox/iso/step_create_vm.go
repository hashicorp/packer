package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/packer"
	"time"
)

// This step creates the actual virtual machine.
//
// Produces:
//   vmName string - The name of the VM
type stepCreateVM struct {
	vmName string
}

func (s *stepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	name := config.VMName

	commands := make([][]string, 4)
	commands[0] = []string{
		"createvm", "--name", name,
		"--ostype", config.GuestOSType, "--register",
	}
	commands[1] = []string{
		"modifyvm", name,
		"--boot1", "disk", "--boot2", "dvd", "--boot3", "none", "--boot4", "none",
	}
	commands[2] = []string{"modifyvm", name, "--cpus", "1"}
	commands[3] = []string{"modifyvm", name, "--memory", "512"}

	ui.Say("Creating virtual machine...")
	for _, command := range commands {
		err := driver.VBoxManage(command...)
		if err != nil {
			err := fmt.Errorf("Error creating VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Set the VM name property on the first command
		if s.vmName == "" {
			s.vmName = name
		}
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.vmName)

	return multistep.ActionContinue
}

func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering and deleting virtual machine...")
	var err error = nil
	for i := 0; i < 5; i++ {
		err = driver.VBoxManage("unregistervm", s.vmName, "--delete")
		if err == nil {
			break
		}

		time.Sleep(1 * time.Second * time.Duration(i))
	}

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
