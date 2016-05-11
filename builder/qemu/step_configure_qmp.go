package qemu

import (
	"fmt"
	"log"
	"net"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step configures the VM to enable the QMP server.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   qmp_port uint - The port that QMP is configured to listen on.
type stepConfigureQMP struct{}

func (stepConfigureQMP) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Find an open QMP port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	msg := fmt.Sprintf("Looking for available port for QMP")
	ui.Say(msg)
	log.Printf(msg)
	l, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		return multistep.ActionHalt
	}
	defer l.Close()
	var qmpPort uint
	qmpPort = uint(l.Addr().(*net.TCPAddr).Port)
	ui.Say(fmt.Sprintf("Found available QMP port: %d", qmpPort))
	state.Put("qmp_port", qmpPort)

	return multistep.ActionContinue
}

func (stepConfigureQMP) Cleanup(multistep.StateBag) {}
