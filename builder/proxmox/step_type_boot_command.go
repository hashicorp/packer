package proxmox

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
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
	MonitorCmd(*proxmox.VmRef, string) (map[string]interface{}, error)
}

var _ commandTyper = &proxmox.Client{}

func (s *stepTypeBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)
	client := state.Get("proxmoxClient").(commandTyper)
	vmRef := state.Get("vmRef").(*proxmox.VmRef)

	if len(s.BootCommand) == 0 {
		log.Println("No boot command given, skipping")
		return multistep.ActionContinue
	}

	if int64(s.BootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot", s.BootWait.String()))
		select {
		case <-time.After(s.BootWait):
			break
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}

	httpIP, err := hostIP()
	if err != nil {
		err := fmt.Errorf("Failed to determine host IP: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	common.SetHTTPIP(httpIP)
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

func (*stepTypeBootCommand) Cleanup(multistep.StateBag) {
	commonhelper.RemoveSharedStateFile("ip", "")
}

func hostIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
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
