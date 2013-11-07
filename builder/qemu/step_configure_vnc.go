package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"math/rand"
	"net"
)

// This step configures the VM to enable the VNC server.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   vnc_port uint - The port that VNC is configured to listen on.
type stepConfigureVNC struct{}

func (stepConfigureVNC) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	msg := fmt.Sprintf("Looking for available port between %d and %d", config.VNCPortMin, config.VNCPortMax)
	ui.Say(msg)
	log.Printf(msg)
	var vncPort uint
	portRange := int(config.VNCPortMax - config.VNCPortMin)
	for {
		vncPort = uint(rand.Intn(portRange)) + config.VNCPortMin
		log.Printf("Trying port: %d", vncPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", vncPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	ui.Say(fmt.Sprintf("Found available VNC port: %d", vncPort))
	state.Put("vnc_port", vncPort)

	return multistep.ActionContinue
}

func (stepConfigureVNC) Cleanup(multistep.StateBag) {}
