package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"math/rand"
	"net"
)

// This step adds a NAT port forwarding definition so that SSH is available
// on the guest machine.
//
// Uses:
//
// Produces:
type stepForwardSSH struct{}

func (s *stepForwardSSH) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	log.Printf("Looking for available SSH port between %d and %d", config.SSHHostPortMin, config.SSHHostPortMax)
	var sshHostPort uint
	portRange := int(config.SSHHostPortMax - config.SSHHostPortMin)
	for {
		sshHostPort = uint(rand.Intn(portRange)) + config.SSHHostPortMin
		log.Printf("Trying port: %d", sshHostPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", sshHostPort))
		if err == nil {
			defer l.Close()
			break
		}
	}
	ui.Say(fmt.Sprintf("Found port for SSH: %d.", sshHostPort))

	// Save the port we're using so that future steps can use it
	state.Put("sshHostPort", sshHostPort)

	return multistep.ActionContinue
}

func (s *stepForwardSSH) Cleanup(state multistep.StateBag) {}
