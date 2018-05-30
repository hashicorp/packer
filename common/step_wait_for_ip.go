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
	SettleTimeout string `mapstructure:"ip_settle_timeout"`

	settleTimeout time.Duration
}

type StepWaitForIp struct {
	Config *WaitIpConfig
}

func (c *WaitIpConfig) Prepare() []error {
	var errs []error

	if c.SettleTimeout == "" {
		c.SettleTimeout = "5s"
	}

	var err error
	c.settleTimeout, err = time.ParseDuration(c.SettleTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed parsing ip_settle_timeout: %s", err))
	}

	return errs
}

func (s *StepWaitForIp) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	ui.Say("Waiting for IP...")

	ipChan := make(chan string)
	errChan := make(chan error)
	go func() {
		doGetIp(vm, ctx, s.Config, errChan, ipChan)
	}()

	for {
		select {
		case err := <-errChan:
			state.Put("error", err)
			return multistep.ActionHalt
		case <-ctx.Done():
			return multistep.ActionHalt
		case ip := <-ipChan:
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

func doGetIp(vm *driver.VirtualMachine, ctx context.Context, c *WaitIpConfig, errChan chan error, ipChan chan string) {
	var prevIp = ""
	var stopTime time.Time
	var interval time.Duration
	if c.settleTimeout.Seconds() >= 120 {
		interval = 30 * time.Second
	} else if c.settleTimeout.Seconds() >= 60 {
		interval = 15 * time.Second
	} else if c.settleTimeout.Seconds() >= 10 {
		interval = 5 * time.Second
	} else {
		interval = 1 * time.Second
	}
loop:
	ip, err := vm.WaitForIP(ctx)
	if err != nil {
		errChan <- err
		return
	}
	if prevIp == "" || prevIp != ip {
		if prevIp == "" {
			log.Printf("VM IP aquired: %s", ip)
		} else {
			log.Printf("VM IP changed from %s to %s", prevIp, ip)
		}
		prevIp = ip
		stopTime = time.Now().Add(c.settleTimeout)
		goto loop
	} else {
		log.Printf("VM IP is still the same: %s", prevIp)
		if time.Now().After(stopTime) {
			log.Printf("VM IP seems stable enough: %s", ip)
			ipChan <- ip
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			goto loop
		}
	}

}

func (s *StepWaitForIp) Cleanup(state multistep.StateBag) {}
