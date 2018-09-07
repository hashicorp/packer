package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort uint
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
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)
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

	hostIp, err := driver.GetHostAdapterIpAddressForSwitch(s.SwitchName)

	if err != nil {
		err := fmt.Errorf("Error getting host adapter ip address: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Host IP for the HyperV machine: %s", hostIp))

	common.SetHTTPIP(hostIp)
	s.Ctx.Data = &bootCommandTemplateData{
		hostIp,
		httpPort,
		vmName,
	}

	sendCodes := func(codes []string) error {
		scanCodesToSendString := strings.Join(codes, " ")
		return driver.TypeScanCodes(vmName, scanCodesToSendString)
	}
	d := bootcommand.NewPCXTDriver(sendCodes, -1, s.GroupInterval)

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
