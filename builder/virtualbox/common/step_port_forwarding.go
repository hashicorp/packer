package common

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/net"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step adds a NAT port forwarding definition so that SSH or WinRM is available
// on the guest machine.
//
// Uses:
//   driver Driver
//   ui packersdk.Ui
//   vmName string
//
// Produces:
type StepPortForwarding struct {
	CommConfig     *communicator.Config
	HostPortMin    int
	HostPortMax    int
	SkipNatMapping bool

	l *net.Listener
}

func (s *StepPortForwarding) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	if s.CommConfig.Type == "none" {
		log.Printf("Not using a communicator, skipping setting up port forwarding...")
		state.Put("commHostPort", 0)
		return multistep.ActionContinue
	}

	guestPort := s.CommConfig.Port()
	commHostPort := guestPort
	if !s.SkipNatMapping {
		log.Printf("Looking for available communicator (SSH, WinRM, etc) port between %d and %d",
			s.HostPortMin, s.HostPortMax)

		var err error
		s.l, err = net.ListenRangeConfig{
			Addr:    "127.0.0.1",
			Min:     s.HostPortMin,
			Max:     s.HostPortMax,
			Network: "tcp",
		}.Listen(ctx)
		if err != nil {
			err := fmt.Errorf("Error creating port forwarding rule: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		s.l.Listener.Close() // free port, but don't unlock lock file
		commHostPort = s.l.Port

		// Make sure to configure the network interface to NAT
		command := []string{
			"modifyvm", vmName,
			"--nic1",
			"nat",
		}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Failed to configure NAT interface: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Create a forwarded port mapping to the VM
		ui.Say(fmt.Sprintf("Creating forwarded port mapping for communicator (SSH, WinRM, etc) (host port %d)", commHostPort))
		command = []string{
			"modifyvm", vmName,
			"--natpf1",
			fmt.Sprintf("packercomm,tcp,127.0.0.1,%d,,%d", commHostPort, guestPort),
		}
		retried := false
	retry:
		if err := driver.VBoxManage(command...); err != nil {
			if !strings.Contains(err.Error(), "A NAT rule of this name already exists") || retried {
				err := fmt.Errorf("Error creating port forwarding rule: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			} else {
				log.Printf("A packer NAT rule already exists. Trying to delete ...")
				delcommand := []string{
					"modifyvm", vmName,
					"--natpf1",
					"delete", "packercomm",
				}
				if err := driver.VBoxManage(delcommand...); err != nil {
					err := fmt.Errorf("Error deleting packer NAT forwarding rule: %s", err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
				goto retry
			}
		}
	}

	// Save the port we're using so that future steps can use it
	state.Put("commHostPort", commHostPort)

	return multistep.ActionContinue
}

func (s *StepPortForwarding) Cleanup(state multistep.StateBag) {
	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			log.Printf("failed to unlock port lockfile: %v", err)
		}
	}
}
