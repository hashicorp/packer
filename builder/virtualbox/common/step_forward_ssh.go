package common

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
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
type StepForwardSSH struct {
	GuestPort   uint
	HostPortMin uint
	HostPortMax uint
}

func (s *StepForwardSSH) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	log.Printf("Looking for available SSH port between %d and %d",
		s.HostPortMin, s.HostPortMax)
	var sshHostPort uint
	var offset uint = 0

	portRange := int(s.HostPortMax - s.HostPortMin)
	if portRange > 0 {
		// Have to check if > 0 to avoid a panic
		offset = uint(rand.Intn(portRange))
	}

	for {
		sshHostPort = offset + s.HostPortMin
		log.Printf("Trying port: %d", sshHostPort)
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", sshHostPort))
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
		fmt.Sprintf("packerssh,tcp,127.0.0.1,%d,,%d", sshHostPort, s.GuestPort),
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

func (s *StepForwardSSH) Cleanup(state multistep.StateBag) {}
