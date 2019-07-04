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

	msg := fmt.Sprintf("QMP socket at: %s", config.QMPSocketPath)
	ui.Say(msg)
	log.Print(msg)

	// Only initialize and open QMP when we have a use for it.
	// Handles cases where user may want the socket, but we don't
	if config.VNCUsePassword {
		// Open QMP socket
		var err error
		var cmd []byte
		var result []byte
		s.monitor, err = qmp.NewSocketMonitor("unix", config.QMPSocketPath, 2*time.Second)
		if err != nil {
			err := fmt.Errorf("Error opening QMP socket: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		QMPMonitor := s.monitor
		vncPassword := state.Get("vnc_password")

		// Connect to QMP
		// function automatically calls capabilities so is immediately ready for commands
		err = QMPMonitor.Connect()
		if err != nil {
			err := fmt.Errorf("Error connecting to QMP socket: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		log.Printf("QMP socket open SUCCESS")

		cmd = []byte(fmt.Sprintf("{ \"execute\": \"change-vnc-password\", \"arguments\": { \"password\": \"%s\" } }",
			vncPassword))
		result, err = QMPMonitor.Run(cmd)
		if err != nil {
			err := fmt.Errorf("Error connecting to QMP socket: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		msg = fmt.Sprintf("QMP Command: %s\nResult: %s", cmd, result)
		log.Printf(msg)

		// Put QMP monitor in statebag in case there is a use in a following step
		// Uncomment for future case as it is unused for now
		//state.Put("qmp_monitor", QMPMonitor)
	}

	return multistep.ActionContinue
}

func (s *stepConfigureQMP) Cleanup(multistep.StateBag) {
	if s.monitor != nil {
		err := s.monitor.Disconnect()
		if err != nil {
			log.Printf("failed to disconnect QMP: %v", err)
		}
	}
}
