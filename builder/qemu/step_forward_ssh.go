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
	var offset uint = 0

	portRange := int(config.SSHHostPortMax - config.SSHHostPortMin)
	if portRange > 0 {
		// Have to check if > 0 to avoid a panic
		offset = uint(rand.Intn(portRange))
	}

	for {
		sshHostPort = offset + config.SSHHostPortMin
		if sshHostPort >= config.SSHHostPortMax {
			offset = 0
			sshHostPort = config.SSHHostPortMin
		}
		log.Printf("Trying port: %d", sshHostPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", sshHostPort))
		if err == nil {
			defer l.Close()
			break
		}
		offset++
	}
	ui.Say(fmt.Sprintf("Found port for communicator (SSH, WinRM, etc): %d.", sshHostPort))

	// Save the port we're using so that future steps can use it
	state.Put("sshHostPort", sshHostPort)

	return multistep.ActionContinue
}

func (s *stepForwardSSH) Cleanup(state multistep.StateBag) {}
