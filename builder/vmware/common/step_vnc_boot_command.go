package common

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/packer/common/bootcommand"
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
	ui := state.Get("ui").(packer.Ui)
	vncIp := state.Get("vnc_ip").(string)
	vncPort := state.Get("vnc_port").(int)
	vncPassword := state.Get("vnc_password")

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

	hostIP := state.Get("http_ip").(string)
	s.Ctx.Data = &VNCBootCommandTemplateData{
		HTTPIP:   hostIP,
		HTTPPort: httpPort,
		Name:     s.VMName,
	}

	d := bootcommand.NewVNCDriver(c, s.Config.BootKeyInterval)

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
