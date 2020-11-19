//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type WaitIpConfig

package common

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type WaitIpConfig struct {
	// Amount of time to wait for VM's IP, similar to 'ssh_timeout'.
	// Defaults to 30m (30 minutes). See the Golang
	// [ParseDuration](https://golang.org/pkg/time/#ParseDuration) documentation
	// for full details.
	WaitTimeout time.Duration `mapstructure:"ip_wait_timeout"`
	// Amount of time to wait for VM's IP to settle down, sometimes VM may
	// report incorrect IP initially, then its recommended to set that
	// parameter to apx. 2 minutes. Examples 45s and 10m. Defaults to
	// 5s(5 seconds). See the Golang
	// [ParseDuration](https://golang.org/pkg/time/#ParseDuration) documentation
	//  for full details.
	SettleTimeout time.Duration `mapstructure:"ip_settle_timeout"`
	// Set this to a CIDR address to cause the service to wait for an address that is contained in
	// this network range. Defaults to "0.0.0.0/0" for any ipv4 address. Examples include:
	//
	// * empty string ("") - remove all filters
	// * `0:0:0:0:0:0:0:0/0` - allow only ipv6 addresses
	// * `192.168.1.0/24` - only allow ipv4 addresses from 192.168.1.1 to 192.168.1.254
	WaitAddress *string `mapstructure:"ip_wait_address"`
	ipnet       *net.IPNet

	// WaitTimeout is a total timeout, so even if VM changes IP frequently and it doesn't settle down we will end waiting.
}

type StepWaitForIp struct {
	Config *WaitIpConfig
}

func (c *WaitIpConfig) Prepare() []error {
	var errs []error

	if c.SettleTimeout == 0 {
		c.SettleTimeout = 5 * time.Second
	}
	if c.WaitTimeout == 0 {
		c.WaitTimeout = 30 * time.Minute
	}
	if c.WaitAddress == nil {
		addr := "0.0.0.0/0"
		c.WaitAddress = &addr
	}

	if *c.WaitAddress != "" {
		var err error
		_, c.ipnet, err = net.ParseCIDR(*c.WaitAddress)
		if err != nil {
			errs = append(errs, fmt.Errorf("unable to parse \"ip_wait_address\": %w", err))
		}
	}

	return errs
}

func (c *WaitIpConfig) GetIPNet() *net.IPNet {
	return c.ipnet
}

func (s *StepWaitForIp) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	var ip string
	var err error

	sub, cancel := context.WithCancel(ctx)
	waitDone := make(chan bool, 1)
	defer func() {
		cancel()
	}()

	go func() {
		ui.Say("Waiting for IP...")
		ip, err = doGetIp(vm, sub, s.Config)
		waitDone <- true
	}()

	log.Printf("[INFO] Waiting for IP, up to total timeout: %s, settle timeout: %s", s.Config.WaitTimeout, s.Config.SettleTimeout)
	timeout := time.After(s.Config.WaitTimeout)
	for {
		select {
		case <-timeout:
			cancel()
			<-waitDone
			if ip != "" {
				state.Put("ip", ip)
				log.Printf("[WARN] API timeout waiting for IP but one IP was found. Using IP: %s", ip)
				return multistep.ActionContinue
			}
			err := fmt.Errorf("Timeout waiting for IP.")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		case <-ctx.Done():
			cancel()
			log.Println("[WARN] Interrupt detected, quitting waiting for IP.")
			return multistep.ActionHalt
		case <-waitDone:
			if err != nil {
				state.Put("error", err)
				return multistep.ActionHalt
			}
			state.Put("ip", ip)
			ui.Say(fmt.Sprintf("IP address: %v", ip))
			return multistep.ActionContinue
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				return multistep.ActionHalt
			}
		}
	}
}

func doGetIp(vm *driver.VirtualMachineDriver, ctx context.Context, c *WaitIpConfig) (string, error) {
	var prevIp = ""
	var stopTime time.Time
	var interval time.Duration
	if c.SettleTimeout.Seconds() >= 120 {
		interval = 30 * time.Second
	} else if c.SettleTimeout.Seconds() >= 60 {
		interval = 15 * time.Second
	} else if c.SettleTimeout.Seconds() >= 10 {
		interval = 5 * time.Second
	} else {
		interval = 1 * time.Second
	}
loop:
	ip, err := vm.WaitForIP(ctx, c.ipnet)
	if err != nil {
		return "", err
	}

	// Check for ctx cancellation to avoid printing any IP logs at the timeout
	select {
	case <-ctx.Done():
		return ip, fmt.Errorf("IP wait cancelled.")
	default:
	}

	if prevIp == "" || prevIp != ip {
		if prevIp == "" {
			log.Printf("VM IP aquired: %s", ip)
		} else {
			log.Printf("VM IP changed from %s to %s", prevIp, ip)
		}
		prevIp = ip
		stopTime = time.Now().Add(c.SettleTimeout)
		goto loop
	} else {
		log.Printf("VM IP is still the same: %s", prevIp)
		if time.Now().After(stopTime) {
			log.Printf("VM IP seems stable enough: %s", ip)
			return ip, nil
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("IP wait cancelled")
		case <-time.After(interval):
			goto loop
		}
	}

}

func (s *StepWaitForIp) Cleanup(state multistep.StateBag) {}
