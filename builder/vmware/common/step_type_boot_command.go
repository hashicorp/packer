package common

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/boot_command"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/go-vnc"
)

// This step "types" the boot command into the VM over VNC.
//
// Uses:
//   http_port int
//   ui     packer.Ui
//   vnc_port uint
//
// Produces:
//   <nothing>
type StepTypeBootCommand struct {
	BootCommand []string
	VNCEnabled  bool
	BootWait    time.Duration
	VMName      string
	Ctx         interpolate.Context
}
type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort uint
	Name     string
}

func (s *StepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.VNCEnabled {
		log.Println("Skipping boot command step...")
		return multistep.ActionContinue
	}

	debug := state.Get("debug").(bool)
	driver := state.Get("driver").(Driver)
	httpPort := state.Get("http_port").(uint)
	ui := state.Get("ui").(packer.Ui)
	vncIp := state.Get("vnc_ip").(string)
	vncPort := state.Get("vnc_port").(uint)
	vncPassword := state.Get("vnc_password")

	// ----------------
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
	// ----------------

	var pauseFn multistep.DebugPauseFn
	if debug {
		pauseFn = state.Get("pauseFn").(multistep.DebugPauseFn)
	}

	// Connect to VNC
	ui.Say(fmt.Sprintf("Connecting to VM via VNC (%s:%d)", vncIp, vncPort))
	nc, err := net.Dial("tcp", fmt.Sprintf("%s:%d", vncIp, vncPort))
	if err != nil {
		err := fmt.Errorf("Error connecting to VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer nc.Close()

	var auth []vnc.ClientAuth

	if vncPassword != nil && len(vncPassword.(string)) > 0 {
		auth = []vnc.ClientAuth{&vnc.PasswordAuth{Password: vncPassword.(string)}}
	} else {
		auth = []vnc.ClientAuth{new(vnc.ClientAuthNone)}
	}

	c, err := vnc.Client(nc, &vnc.ClientConfig{Auth: auth, Exclusive: true})
	if err != nil {
		err := fmt.Errorf("Error handshaking with VNC: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer c.Close()

	log.Printf("Connected to VNC desktop: %s", c.DesktopName)

	// Determine the host IP
	hostIP, err := driver.HostIP(state)
	if err != nil {
		err := fmt.Errorf("Error detecting host IP: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Host IP for the VMware machine: %s", hostIP)
	common.SetHTTPIP(hostIP)

	s.Ctx.Data = &bootCommandTemplateData{
		hostIP,
		httpPort,
		s.VMName,
	}

	d := bootcommand.NewVNCDriver(c)

	ui.Say("Typing the boot command over VNC...")
	for i, command := range s.BootCommand {
		command, err := interpolate.Render(command, &s.Ctx)
		if err != nil {
			err := fmt.Errorf("Error preparing boot command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if pauseFn != nil {
			pauseFn(multistep.DebugLocationAfterRun, fmt.Sprintf("boot_command[%d]: %s", i, command), state)
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
	}

	return multistep.ActionContinue
}

func (*StepTypeBootCommand) Cleanup(multistep.StateBag) {}
