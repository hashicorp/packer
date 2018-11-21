package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type commandTemplate struct {
	Name string
}

// This step executes additional VBoxManage commands as specified by the
// template.
//
// Uses:
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepVBoxManage struct {
	Commands []string
	Ctx      interpolate.Context
}

func (s *StepVBoxManage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if len(s.Commands) > 0 {
		ui.Say("Executing custom VBoxManage commands...")
	}

	s.Ctx.Data = &commandTemplate{
		Name: vmName,
	}

	commands, err := s.Ctx.ParseArgs(s.Commands)
	if err != nil {
		err = fmt.Errorf("Error preparing vboxmanage command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, command := range commands {

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

func (s *StepVBoxManage) Cleanup(state multistep.StateBag) {}
