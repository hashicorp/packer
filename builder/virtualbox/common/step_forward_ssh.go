package common

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
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
	HostPortMin    uint
	HostPortMax    uint
	SkipNatMapping bool
}

func (s *StepForwardSSH) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	guestPort := s.CommConfig.Port()
	sshHostPort := guestPort
	if !s.SkipNatMapping {
		log.Printf("Looking for available communicator (SSH, WinRM, etc) port between %d and %d",
			s.HostPortMin, s.HostPortMax)
		offset := 0

		portRange := int(s.HostPortMax - s.HostPortMin)
		if portRange > 0 {
			// Have to check if > 0 to avoid a panic
			offset = rand.Intn(portRange)
		}

		for {
			sshHostPort = offset + int(s.HostPortMin)
			if sshHostPort >= int(s.HostPortMax) {
				offset = 0
				sshHostPort = int(s.HostPortMin)
			}
			log.Printf("Trying port: %d", sshHostPort)
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", sshHostPort))
			if err == nil {
				defer l.Close()
				break
			}
			offset++
		}

		// Create a forwarded port mapping to the VM
		ui.Say(fmt.Sprintf("Creating forwarded port mapping for communicator (SSH, WinRM, etc) (host port %d)", sshHostPort))
		command := []string{
			"modifyvm", vmName,
			"--natpf1",
			fmt.Sprintf("packercomm,tcp,127.0.0.1,%d,,%d", sshHostPort, guestPort),
		}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error creating port forwarding rule: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Save the port we're using so that future steps can use it
	state.Put("sshHostPort", sshHostPort)

	return multistep.ActionContinue
}

func (s *StepForwardSSH) Cleanup(state multistep.StateBag) {}
