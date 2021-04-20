package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type commandTemplate struct {
	Name string
}

// StepPrlctl is a step that executes additional `prlctl` commands as specified.
// by the template.
//
// Uses:
//   driver Driver
//   ui packersdk.Ui
//   vmName string
//
// Produces:
type StepPrlctl struct {
	Commands [][]string
	Ctx      interpolate.Context
}

// Run executes `prlctl` commands.
func (s *StepPrlctl) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	if len(s.Commands) > 0 {
		ui.Say("Executing custom prlctl commands...")
	}

	s.Ctx.Data = &commandTemplate{
		Name: vmName,
	}

	for _, originalCommand := range s.Commands {
		command := make([]string, len(originalCommand))
		copy(command, originalCommand)

		for i, arg := range command {
			var err error
			command[i], err = interpolate.Render(arg, &s.Ctx)
			if err != nil {
				err = fmt.Errorf("Error preparing prlctl command: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}

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
