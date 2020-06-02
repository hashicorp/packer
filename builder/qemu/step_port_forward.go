package qemu

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step adds a NAT port forwarding definition so that SSH or WinRM is available
// on the guest machine.
type stepPortForward struct {
	l *net.Listener
}

func (s *stepPortForward) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	commHostPort := config.CommConfig.Comm.Port()

	if config.CommConfig.SkipNatMapping {
		log.Printf("Skipping NAT port forwarding. Using communicator (SSH, WinRM, etc) port %d", commHostPort)
		state.Put("commHostPort", commHostPort)
		return multistep.ActionContinue
	}

	log.Printf("Looking for available communicator (SSH, WinRM, etc) port between %d and %d", config.CommConfig.HostPortMin, config.CommConfig.HostPortMax)
	var err error
	s.l, err = net.ListenRangeConfig{
		Addr:    config.VNCBindAddress,
		Min:     config.CommConfig.HostPortMin,
		Max:     config.CommConfig.HostPortMax,
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		err := fmt.Errorf("Error finding port: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.l.Listener.Close() // free port, but don't unlock lock file
	commHostPort = s.l.Port
	ui.Say(fmt.Sprintf("Found port for communicator (SSH, WinRM, etc): %d.", commHostPort))

	// Save the port we're using so that future steps can use it
	state.Put("commHostPort", commHostPort)

	return multistep.ActionContinue
}

func (s *stepPortForward) Cleanup(state multistep.StateBag) {
	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			log.Printf("failed to unlock port lockfile: %v", err)
		}
	}
}
