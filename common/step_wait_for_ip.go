package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"log"
	"time"
)

type WaitIpConfig struct {
	WaitTimeout   time.Duration `mapstructure:"ip_wait_timeout"`
	SettleTimeout time.Duration `mapstructure:"ip_settle_timeout"`

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

	return errs
}

func (s *StepWaitForIp) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

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
			err := fmt.Errorf("Timeout waiting for IP.")
			state.Put("error", err)
			ui.Error(err.Error())
			cancel()
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

func doGetIp(vm *driver.VirtualMachine, ctx context.Context, c *WaitIpConfig) (string, error) {
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
	ip, err := vm.WaitForIP(ctx)
	if err != nil {
		return "", err
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
