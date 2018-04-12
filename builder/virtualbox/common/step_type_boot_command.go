package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/boot_command"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const KeyLeftShift uint32 = 0xFFE1

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort uint
	Name     string
}

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   driver Driver
//   http_port int
//   ui     packer.Ui
//   vmName string
//
// Produces:
//   <nothing>
type StepTypeBootCommand struct {
	BootCommand []string
	BootWait    time.Duration
	VMName      string
	Ctx         interpolate.Context
}

func (s *StepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	debug := state.Get("debug").(bool)
	driver := state.Get("driver").(Driver)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Wait the for the vm to boot.
	if int64(s.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", s.BootWait.String()))
		select {
		case <-time.After(s.BootWait):
			break
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}

	var pauseFn multistep.DebugPauseFn
	if debug {
		pauseFn = state.Get("pauseFn").(multistep.DebugPauseFn)
	}

	hostIP := "10.0.2.2"
	common.SetHTTPIP(hostIP)
	s.Ctx.Data = &bootCommandTemplateData{
		hostIP,
		httpPort,
		s.VMName,
	}

	sendCodes := func(codes []string) error {
		args := []string{"controlvm", vmName, "keyboardputscancode"}
		args = append(args, codes...)

		if err := driver.VBoxManage(args...); err != nil {
			return err
		}
		return nil
	}
	d := bootcommand.NewPCATDriver(sendCodes)

	ui.Say("Typing the boot command...")
	for i, command := range s.BootCommand {
		command, err := interpolate.Render(command, &s.Ctx)
		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		seq, err := bootcommand.GenerateExpressionSequence(command)
		if err != nil {
			err := fmt.Errorf("Error generating boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// This executes vboxmanage once for each character code. This seems
		// fine for now, but changes the prior behavior. If this becomes
		// a problem, we can always have the driver cache scancodes, and then
		// add a `Flush` method which we can call after this.
		if err := seq.Do(ctx, d); err != nil {
			err := fmt.Errorf("Error running boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if pauseFn != nil {
			pauseFn(multistep.DebugLocationAfterRun, fmt.Sprintf("boot_command[%d]: %s", i, command), state)
		}

	}

	return multistep.ActionContinue
}

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}
