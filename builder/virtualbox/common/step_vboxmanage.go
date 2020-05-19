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
	// HTTPIP is the HTTP server's IP address.
	HTTPIP string

	// HTTPPort is the HTTP server port.
	HTTPPort int

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
	Commands [][]string
	Ctx      interpolate.Context
}

func (s *StepVBoxManage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if len(s.Commands) > 0 {
		ui.Say("Executing custom VBoxManage commands...")
	}

	hostIP := state.Get("http_ip").(string)
	httpPort := state.Get("http_port").(int)

	s.Ctx.Data = &commandTemplate{
		Name:     vmName,
		HTTPIP:   hostIP,
		HTTPPort: httpPort,
	}

	for _, originalCommand := range s.Commands {
		command := make([]string, len(originalCommand))
		copy(command, originalCommand)

		for i, arg := range command {
			var err error
			command[i], err = interpolate.Render(arg, &s.Ctx)
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

func (s *StepVBoxManage) Cleanup(state multistep.StateBag) {}
