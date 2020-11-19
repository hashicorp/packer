package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/net"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step configures the VM to enable the VRDP server
// on the guest machine.
//
// Uses:
//   driver Driver
//   ui packersdk.Ui
//   vmName string
//
// Produces:
// vrdp_port unit - The port that VRDP is configured to listen on.
type StepConfigureVRDP struct {
	VRDPBindAddress string
	VRDPPortMin     int
	VRDPPortMax     int

	l *net.Listener
}

func (s *StepConfigureVRDP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	log.Printf("Looking for available port between %d and %d on %s", s.VRDPPortMin, s.VRDPPortMax, s.VRDPBindAddress)
	var err error
	s.l, err = net.ListenRangeConfig{
		Addr:    s.VRDPBindAddress,
		Min:     s.VRDPPortMin,
		Max:     s.VRDPPortMax,
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		err := fmt.Errorf("Error finding port: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.l.Listener.Close() // free port, but don't unlock lock file
	vrdpPort := s.l.Port

	command := []string{
		"modifyvm", vmName,
		"--vrdeaddress", s.VRDPBindAddress,
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

	state.Put("vrdpIp", s.VRDPBindAddress)
	state.Put("vrdpPort", vrdpPort)

	return multistep.ActionContinue
}

func (s *StepConfigureVRDP) Cleanup(state multistep.StateBag) {
	if s.l != nil {
		err := s.l.Close()
		if err != nil {
			log.Printf("failed to unlock port lockfile: %v", err)
		}
	}
}
