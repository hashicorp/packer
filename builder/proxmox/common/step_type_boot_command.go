package proxmox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/packer-plugin-sdk/bootcommand"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

// stepTypeBootCommand takes the started VM, and sends the keystrokes required to start
// the installation process such that Packer can later reach the VM over SSH/WinRM
type stepTypeBootCommand struct {
	bootcommand.BootConfig
	Ctx interpolate.Context
}

type bootCommandTemplateData struct {
	HTTPIP   string
	HTTPPort int
}

type commandTyper interface {
	Sendkey(*proxmox.VmRef, string) error
}

var _ commandTyper = &proxmox.Client{}

func (s *stepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	client := state.Get("proxmoxClient").(commandTyper)
	vmRef := state.Get("vmRef").(*proxmox.VmRef)

	if len(s.BootCommand) == 0 {
		log.Println("No boot command given, skipping")
		return multistep.ActionContinue
	}

	if int64(s.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot", s.BootWait))
		select {
		case <-time.After(s.BootWait):
			break
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}
	var httpIP string
	var err error
	if c.HTTPAddress != "0.0.0.0" {
		httpIP = c.HTTPAddress
	} else {
		httpIP, err = hostIP(c.HTTPInterface)
		if err != nil {
			err := fmt.Errorf("Failed to determine host IP: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	state.Put("http_ip", httpIP)
	s.Ctx.Data = &bootCommandTemplateData{
		HTTPIP:   httpIP,
		HTTPPort: state.Get("http_port").(int),
	}

	ui.Say("Typing the boot command")
	d := NewProxmoxDriver(client, vmRef, c.BootKeyInterval)
	command, err := interpolate.Render(s.FlatBootCommand(), &s.Ctx)
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

func (*stepTypeBootCommand) Cleanup(multistep.StateBag) {}

func hostIP(ifname string) (string, error) {
	var addrs []net.Addr
	var err error

	if ifname != "" {
		iface, err := net.InterfaceByName(ifname)
		if err != nil {
			return "", err
		}
		addrs, err = iface.Addrs()
		if err != nil {
			return "", err
		}
	} else {
		addrs, err = net.InterfaceAddrs()
		if err != nil {
			return "", err
		}
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("No host IP found")
}
