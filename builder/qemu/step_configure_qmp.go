package qemu

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/digitalocean/go-qemu/qmp"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// This step configures the VM to enable the QMP listener.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
type stepConfigureQMP struct {
	monitor *qmp.SocketMonitor
}

func (s *stepConfigureQMP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if !config.QMPEnable {
		return multistep.ActionContinue
	}

	msg := fmt.Sprintf("Opening QMP socket at: %s", config.QMPSocketPath)
	ui.Say(msg)
	log.Print(msg)

	// Open QMP socket
	var err error
	s.monitor, err = qmp.NewSocketMonitor("unix", config.QMPSocketPath, 2*time.Second)
	if err != nil {
		err := fmt.Errorf("Error opening QMP socket: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	QMPMonitor := s.monitor

	log.Printf("QMP socket open SUCCESS")

	state.Put("qmp_monitor", QMPMonitor)

	return multistep.ActionContinue
}

func (s *stepConfigureQMP) Cleanup(multistep.StateBag) {
}
