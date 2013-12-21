package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
)

type commandTemplate struct {
	Name string
}

// This step executes additional VBoxManage commands as specified by the
// template.
//
// Uses:
//
// Produces:
type stepVBoxManage struct{}

func (s *stepVBoxManage) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if len(config.VBoxManage) > 0 {
		ui.Say("Executing custom VBoxManage commands...")
	}

	tplData := &commandTemplate{
		Name: vmName,
	}

	for _, originalCommand := range config.VBoxManage {
		command := make([]string, len(originalCommand))
		copy(command, originalCommand)

		for i, arg := range command {
			var err error
			command[i], err = config.tpl.Process(arg, tplData)
			if err != nil {
				err := fmt.Errorf("Error preparing vboxmanage command: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}

		ui.Message(fmt.Sprintf("Executing: %s", strings.Join(command, " ")))
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error executing command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepVBoxManage) Cleanup(state multistep.StateBag) {}
