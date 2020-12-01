package qemu

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// Step to discover the http ip
// which guests use to reach the vm host
// To make sure the IP is set before boot command and http server steps
type stepHTTPIPDiscover struct{}

func (s *stepHTTPIPDiscover) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	hostIP := ""

	if config.NetBridge == "" {
		hostIP = "10.0.2.2"
	} else {
		bridgeInterface, err := net.InterfaceByName(config.NetBridge)
		if err != nil {
			err := fmt.Errorf("Error getting the bridge %s interface: %s", config.NetBridge, err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		addrs, err := bridgeInterface.Addrs()
		if err != nil {
			err := fmt.Errorf("Error getting the bridge %s interface addresses: %s", config.NetBridge, err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			hostIP = ip.String()
			break
		}
		if hostIP == "" {
			err := fmt.Errorf("Error getting an IPv4 address from the bridge %s: cannot find any IPv4 address", config.NetBridge)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	state.Put("http_ip", hostIP)

	return multistep.ActionContinue
}

func (s *stepHTTPIPDiscover) Cleanup(state multistep.StateBag) {}
