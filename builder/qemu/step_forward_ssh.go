package qemu

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step adds a NAT port forwarding definition so that SSH is available
// on the guest machine.
//
// Uses:
//
// Produces:
type stepForwardSSH struct {
	l *net.Listener
}

func (s *stepForwardSSH) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	log.Printf("Looking for available communicator (SSH, WinRM, etc) port between %d and %d", config.SSHHostPortMin, config.SSHHostPortMax)
	var err error
	s.l, err = net.ListenRangeConfig{
		Addr:    config.VNCBindAddress,
		Min:     config.SSHHostPortMin,
		Max:     config.SSHHostPortMax,
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		err := fmt.Errorf("Error finding port: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.l.Listener.Close() // free port, but don't unlock lock file
	sshHostPort := s.l.Port
	ui.Say(fmt.Sprintf("Found port for communicator (SSH, WinRM, etc): %d.", sshHostPort))

	// Save the port we're using so that future steps can use it
	state.Put("sshHostPort", sshHostPort)

	return multistep.ActionContinue
}

func (s *stepForwardSSH) Cleanup(state multistep.StateBag) {
	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			log.Printf("failed to unlock port lockfile: %v", err)
		}
	}
}
