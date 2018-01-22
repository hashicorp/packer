package common

import (
	"fmt"
	"log"
	"math/rand"
	"net"

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
	HostPortMin    uint
	HostPortMax    uint
	SkipNatMapping bool
}

func (s *StepForwardSSH) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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

		portRange := int(s.HostPortMax - s.HostPortMin + 1)
		offset := rand.Intn(portRange)

		for {
			sshHostPort = offset + int(s.HostPortMin)
			log.Printf("Trying port: %d", sshHostPort)
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", sshHostPort))
			if err == nil {
				defer l.Close()
				break
			}
			offset++
			if offset == portRange {
				offset = 0
			}
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
