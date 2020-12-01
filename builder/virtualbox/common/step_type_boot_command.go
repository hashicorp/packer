package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

const KeyLeftShift uint32 = 0xFFE1

// TODO: Should this be made available for other builders?
//  It is copy pasted in the VMWare builder as well.
type bootCommandTemplateData struct {
	// HTTPIP is the HTTP server's IP address.
	HTTPIP string

	// HTTPPort is the HTTP server port.
	HTTPPort int

	// Name is the VM's name.
	Name string

	// SSHPublicKey is the SSH public key in OpenSSH authorized_keys format.
	SSHPublicKey string
}

type StepTypeBootCommand struct {
	BootCommand   string
	BootWait      time.Duration
	VMName        string
	Ctx           interpolate.Context
	GroupInterval time.Duration
	Comm          *communicator.Config
}

func (s *StepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	debug := state.Get("debug").(bool)
	driver := state.Get("driver").(Driver)
	httpPort := state.Get("http_port").(int)
	ui := state.Get("ui").(packersdk.Ui)
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

	hostIP := state.Get("http_ip").(string)
	s.Ctx.Data = &bootCommandTemplateData{
		HTTPIP:       hostIP,
		HTTPPort:     httpPort,
		Name:         s.VMName,
		SSHPublicKey: string(s.Comm.SSHPublicKey),
	}

	sendCodes := func(codes []string) error {
		args := []string{"controlvm", vmName, "keyboardputscancode"}
		args = append(args, codes...)

		return driver.VBoxManage(args...)
	}
	d := bootcommand.NewPCXTDriver(sendCodes, 25, s.GroupInterval)

	ui.Say("Typing the boot command...")
	command, err := interpolate.Render(s.BootCommand, &s.Ctx)
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

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}
