package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/go-vnc"
)

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   http_port int
//   ui     packersdk.Ui
//   vnc_port int
//
// Produces:
//   <nothing>
type StepVNCBootCommand struct {
	Config bootcommand.VNCConfig
	VMName string
	Ctx    interpolate.Context
}

type VNCBootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort int
	Name     string
}

func (s *StepVNCBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Config.DisableVNC {
		log.Println("Skipping boot command step...")
		return multistep.ActionContinue
	}

	debug := state.Get("debug").(bool)
	httpPort := state.Get("http_port").(int)
	ui := state.Get("ui").(packersdk.Ui)
	conn := state.Get("vnc_conn").(*vnc.ClientConn)
	defer conn.Close()

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

	hostIP := state.Get("http_ip").(string)
	s.Ctx.Data = &VNCBootCommandTemplateData{
		HTTPIP:   hostIP,
		HTTPPort: httpPort,
		Name:     s.VMName,
	}

	d := bootcommand.NewVNCDriver(conn, s.Config.BootKeyInterval)

	ui.Say("Typing the boot command over VNC...")
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
		pauseFn(multistep.DebugLocationAfterRun,
			fmt.Sprintf("boot_command: %s", command), state)
	}

	return multistep.ActionContinue
}

func (*StepVNCBootCommand) Cleanup(multistep.StateBag) {}
