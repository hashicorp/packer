package common

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step configures the VM to enable the VRDP server
// on the guest machine.
//
// Uses:
//   driver Driver
//   ui packer.Ui
//   vmName string
//
// Produces:
// vrdp_port unit - The port that VRDP is configured to listen on.
type StepConfigureVRDP struct {
	VRDPPortMin uint
	VRDPPortMax uint
}

func (s *StepConfigureVRDP) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	log.Printf("Looking for available port between %d and %d", s.VRDPPortMin, s.VRDPPortMax)
	var vrdpPort uint
	portRange := int(s.VRDPPortMax - s.VRDPPortMin)

	for {
		if portRange > 0 {
			vrdpPort = uint(rand.Intn(portRange)) + s.VRDPPortMin
		} else {
			vrdpPort = s.VRDPPortMin
		}

		log.Printf("Trying port: %d", vrdpPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", vrdpPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	command := []string{
		"modifyvm", vmName,
		"--vrdeaddress", "127.0.0.1",
		"--vrdeauthtype", "null",
		"--vrde", "on",
		"--vrdeport",
		fmt.Sprintf("%d", vrdpPort),
	}
	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error enabling VRDP: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vrdpIp", "127.0.0.1")
	state.Put("vrdpPort", vrdpPort)

	return multistep.ActionContinue
}

func (s *StepConfigureVRDP) Cleanup(state multistep.StateBag) {}
