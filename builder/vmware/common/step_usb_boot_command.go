package common

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"golang.org/x/mobile/event/key"
)

// This step "types" the boot command into the VM using USB Scan Codes.
type StepUSBBootCommand struct {
	Config      bootcommand.BootConfig
	KeyInterval time.Duration
	VMName      string
	Ctx         interpolate.Context
}

type USBBootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort int
	Name     string
}

func (s *StepUSBBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	debug := state.Get("debug").(bool)
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(*ESX5Driver)

	if s.Config.BootCommand == nil {
		return multistep.ActionContinue
	}

	// Wait the for the vm to boot.
	if int64(s.Config.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", s.Config.BootWait.String()))
		select {
		case <-time.After(s.Config.BootWait):
			break
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}

	var pauseFn multistep.DebugPauseFn
	if debug {
		pauseFn = state.Get("pauseFn").(multistep.DebugPauseFn)
	}

	port := state.Get("http_port").(int)
	if port > 0 {
		ip := state.Get("http_ip").(string)
		s.Ctx.Data = &USBBootCommandTemplateData{
			HTTPIP:   ip,
			HTTPPort: port,
			Name:     s.VMName,
		}
		ui.Say(fmt.Sprintf("HTTP server is working at http://%v:%v/", ip, port))
	}

	var keyAlt, keyCtrl, keyShift bool
	sendCodes := func(code key.Code, down bool) error {
		switch code {
		case key.CodeLeftAlt:
			keyAlt = down
		case key.CodeLeftControl:
			keyCtrl = down
		case key.CodeLeftShift:
			keyShift = down
		}

		shift := down
		if keyShift {
			shift = keyShift
		}

		_, err := driver.TypeOnKeyboard(KeyInput{
			Scancode: code,
			Ctrl:     keyCtrl,
			Alt:      keyAlt,
			Shift:    shift,
		})
		if err != nil {
			return fmt.Errorf("error typing a boot command (code, down) `%d, %t`: %w", code, down, err)
		}
		return nil
	}
	d := bootcommand.NewUSBDriver(sendCodes, s.KeyInterval)

	ui.Say("Typing boot command...")
	flatBootCommand := s.Config.FlatBootCommand()
	command, err := interpolate.Render(flatBootCommand, &s.Ctx)
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

func (*StepUSBBootCommand) Cleanup(multistep.StateBag) {}
