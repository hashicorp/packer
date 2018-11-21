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

// StepPrlctl is a step that executes additional `prlctl` commands as specified.
// by the template.
//
// Uses:
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepPrlctl struct {
	Commands []string
	Ctx      interpolate.Context
}

// Run executes `prlctl` commands.
func (s *StepPrlctl) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if len(s.Commands) > 0 {
		ui.Say("Executing custom prlctl commands...")
	}

	s.Ctx.Data = &commandTemplate{
		Name: vmName,
	}

	commands, err := s.Ctx.ParseArgs(s.Commands)
	if err != nil {
		err = fmt.Errorf("Error preparing prlctl command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, command := range commands {
		ui.Message(fmt.Sprintf("Executing: prlctl %s", strings.Join(command, " ")))
		if err := driver.Prlctl(command...); err != nil {
			err = fmt.Errorf("Error executing command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

// Cleanup does nothing.
func (s *StepPrlctl) Cleanup(state multistep.StateBag) {}
