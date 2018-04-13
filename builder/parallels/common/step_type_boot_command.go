package common

import (
	"context"
	"fmt"
	"log"
	"time"

	packer_common "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/boot_command"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort uint
	Name     string
}

// StepTypeBootCommand is a step that "types" the boot command into the VM via
// the prltype script, built on the Parallels Virtualization SDK - Python API.
type StepTypeBootCommand struct {
	BootCommand    []string
	BootWait       time.Duration
	HostInterfaces []string
	VMName         string
	Ctx            interpolate.Context
}

// Run types the boot command by sending key scancodes into the VM.
func (s *StepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	debug := state.Get("debug").(bool)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

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

	hostIP := "0.0.0.0"

	if len(s.HostInterfaces) > 0 {
		// Determine the host IP
		ipFinder := &IfconfigIPFinder{Devices: s.HostInterfaces}

		ip, err := ipFinder.HostIP()
		if err != nil {
			err = fmt.Errorf("Error detecting host IP: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		hostIP = ip
	}

	ui.Say(fmt.Sprintf("Host IP for the Parallels machine: %s", hostIP))

	packer_common.SetHTTPIP(hostIP)
	s.Ctx.Data = &bootCommandTemplateData{
		hostIP,
		httpPort,
		s.VMName,
	}

	sendCodes := func(codes []string) error {
		log.Printf("Sending scancodes: %#v", codes)
		return driver.SendKeyScanCodes(s.VMName, codes...)
	}
	d := bootcommand.NewPCATDriver(sendCodes, -1)

	ui.Say("Typing the boot command...")
	for i, command := range s.BootCommand {
		command, err := interpolate.Render(command, &s.Ctx)
		if err != nil {
			err = fmt.Errorf("Error preparing boot command: %s", err)
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

// Cleanup does nothing.
func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}
