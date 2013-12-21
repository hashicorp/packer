package iso

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
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	log.Printf("Looking for available SSH port between %d and %d", config.SSHHostPortMin, config.SSHHostPortMax)
	var sshHostPort uint
	var offset uint = 0

	portRange := int(config.SSHHostPortMax - config.SSHHostPortMin)
	if portRange > 0 {
		// Have to check if > 0 to avoid a panic
		offset = uint(rand.Intn(portRange))
	}

	for {
		sshHostPort = offset + config.SSHHostPortMin
		log.Printf("Trying port: %d", sshHostPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", sshHostPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	// Create a forwarded port mapping to the VM
	ui.Say(fmt.Sprintf("Creating forwarded port mapping for SSH (host port %d)", sshHostPort))
	command := []string{
		"modifyvm", vmName,
		"--natpf1",
		fmt.Sprintf("packerssh,tcp,127.0.0.1,%d,,%d", sshHostPort, config.SSHPort),
	}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error creating port forwarding rule: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Save the port we're using so that future steps can use it
	state.Put("sshHostPort", sshHostPort)

	return multistep.ActionContinue
}

func (s *stepForwardSSH) Cleanup(state multistep.StateBag) {}
