package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort int
	Name     string
}

// This step "types" the boot command into the VM via the Hyper-V virtual keyboard
type StepTypeBootCommand struct {
	BootCommand   string
	BootWait      time.Duration
	SwitchName    string
	Ctx           interpolate.Context
	GroupInterval time.Duration
}

func (s *StepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	httpPort := state.Get("http_port").(int)
	ui := state.Get("ui").(packersdk.Ui)
	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)
	hostIp := state.Get("http_ip").(string)

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

	s.Ctx.Data = &bootCommandTemplateData{
		hostIp,
		httpPort,
		vmName,
	}

	sendCodes := func(codes []string) error {
		scanCodesToSendString := strings.Join(codes, " ")
		return driver.TypeScanCodes(vmName, scanCodesToSendString)
	}
	d := bootcommand.NewPCXTDriver(sendCodes, 32, s.GroupInterval)

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

	return multistep.ActionContinue
}

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}
