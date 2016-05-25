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
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	msg := fmt.Sprintf("Looking for available port between %d and %d on %s", config.VNCPortMin, config.VNCPortMax, config.VNCBindAddress)
	ui.Say(msg)
	log.Printf(msg)
	var vncPort uint
	portRange := int(config.VNCPortMax - config.VNCPortMin)
	for {
		if portRange > 0 {
			vncPort = uint(rand.Intn(portRange)) + config.VNCPortMin
		} else {
			vncPort = config.VNCPortMin
		}

		log.Printf("Trying port: %d", vncPort)
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.VNCBindAddress, vncPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	log.Printf("Found available VNC port: %d on IP: %s", vncPort, config.VNCBindAddress)
	state.Put("vnc_port", vncPort)
	state.Put("vnc_ip", config.VNCBindAddress)

	return multistep.ActionContinue
}

func (stepConfigureVNC) Cleanup(multistep.StateBag) {}
