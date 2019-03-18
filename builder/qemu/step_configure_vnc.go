package qemu

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step configures the VM to enable the VNC server.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   vnc_port uint - The port that VNC is configured to listen on.
type stepConfigureVNC struct {
	l *net.Listener
}

func (s *stepConfigureVNC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	msg := fmt.Sprintf("Looking for available port between %d and %d on %s", config.VNCPortMin, config.VNCPortMax, config.VNCBindAddress)
	ui.Say(msg)
	log.Print(msg)

	var err error
	s.l, err = net.ListenRangeConfig{
		Addr:    config.VNCBindAddress,
		Min:     config.VNCPortMin,
		Max:     config.VNCPortMax,
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		err := fmt.Errorf("Error finding port: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.l.Listener.Close() // free port, but don't unlock lock file
	vncPort := s.l.Port

	log.Printf("Found available VNC port: %d on IP: %s", vncPort, config.VNCBindAddress)
	state.Put("vnc_port", vncPort)
	state.Put("vnc_ip", config.VNCBindAddress)

	return multistep.ActionContinue
}

func (s *stepConfigureVNC) Cleanup(multistep.StateBag) {
	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			log.Printf("failed to unlock port lockfile: %v", err)
		}
	}
}
