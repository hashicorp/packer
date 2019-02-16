package common

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step adds a NAT port forwarding definition so that SSH is available
// on the guest machine.
//
// Uses:
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepForwardSSH struct {
	CommConfig     *communicator.Config
	HostPortMin    int
	HostPortMax    int
	SkipNatMapping bool

	l *net.Listener
}

func (s *StepForwardSSH) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if s.CommConfig.Type == "none" {
		log.Printf("Not using a communicator, skipping setting up port forwarding...")
		state.Put("sshHostPort", 0)
		return multistep.ActionContinue
	}

	guestPort := s.CommConfig.Port()
	sshHostPort := guestPort
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
		sshHostPort = s.l.Port

		// Create a forwarded port mapping to the VM
		ui.Say(fmt.Sprintf("Creating forwarded port mapping for communicator (SSH, WinRM, etc) (host port %d)", sshHostPort))
		command := []string{
			"modifyvm", vmName,
			"--natpf1",
			fmt.Sprintf("packercomm,tcp,127.0.0.1,%d,,%d", sshHostPort, guestPort),
		}
		if err := driver.VBoxManage(command...); err != nil {
			if !strings.Contains(err.Error(), "A NAT rule of this name already exists") {
				err := fmt.Errorf("Error creating port forwarding rule: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Save the port we're using so that future steps can use it
	state.Put("sshHostPort", sshHostPort)

	return multistep.ActionContinue
}

func (s *StepForwardSSH) Cleanup(state multistep.StateBag) {
	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			log.Printf("failed to unlock port lockfile: %v", err)
		}
	}
}
