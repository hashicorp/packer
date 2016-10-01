package qemu

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step adds a NAT port forwarding definition so that SSH is available
// on the guest machine.
//
// Uses:
//
// Produces:
type stepForwardSSH struct{}

func (s *stepForwardSSH) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	log.Printf("Looking for available communicator (SSH, WinRM, etc) port between %d and %d", config.SSHHostPortMin, config.SSHHostPortMax)
	var sshHostPort uint

	portRange := config.SSHHostPortMax - config.SSHHostPortMin + 1
	offset := uint(rand.Intn(int(portRange)))

	for {
		sshHostPort = offset + config.SSHHostPortMin
		log.Printf("Trying port: %d", sshHostPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", sshHostPort))
		if err == nil {
			defer l.Close()
			break
		}
		offset++
		if offset == portRange {
			offset = 0
		}
	}
	ui.Say(fmt.Sprintf("Found port for communicator (SSH, WinRM, etc): %d.", sshHostPort))

	// Save the port we're using so that future steps can use it
	state.Put("sshHostPort", sshHostPort)

	return multistep.ActionContinue
}

func (s *stepForwardSSH) Cleanup(state multistep.StateBag) {}
