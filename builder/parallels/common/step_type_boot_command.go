package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort int
	Name     string
}

// StepTypeBootCommand is a step that "types" the boot command into the VM via
// the prltype script, built on the Parallels Virtualization SDK - Python API.
type StepTypeBootCommand struct {
	BootCommand    string
	BootWait       time.Duration
	HostInterfaces []string
	VMName         string
	Ctx            interpolate.Context
	GroupInterval  time.Duration
}

// Run types the boot command by sending key scancodes into the VM.
func (s *StepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	debug := state.Get("debug").(bool)
	httpPort := state.Get("http_port").(int)
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

	state.Put("http_ip", hostIP)
	s.Ctx.Data = &bootCommandTemplateData{
		hostIP,
		httpPort,
		s.VMName,
	}

	sendCodes := func(codes []string) error {
		return driver.SendKeyScanCodes(s.VMName, codes...)
	}
	d := bootcommand.NewPCXTDriver(sendCodes, -1, s.GroupInterval)

	ui.Say("Typing the boot command...")
	command, err := interpolate.Render(s.BootCommand, &s.Ctx)
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
		pauseFn(multistep.DebugLocationAfterRun, fmt.Sprintf("boot_command: %s", command), state)
	}

	return multistep.ActionContinue
}

// Cleanup does nothing.
func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}
