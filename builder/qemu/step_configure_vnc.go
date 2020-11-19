package qemu

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/net"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step configures the VM to enable the VNC server.
//
// Uses:
//   config *config
//   ui     packersdk.Ui
//
// Produces:
//   vnc_port int - The port that VNC is configured to listen on.
type stepConfigureVNC struct {
	l *net.Listener
}

func VNCPassword() string {
	length := int(8)

	charSet := []byte("012345689abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	charSetLength := len(charSet)

	password := make([]byte, length)

	for i := 0; i < length; i++ {
		password[i] = charSet[rand.Intn(charSetLength)]
	}

	return string(password)
}

func (s *stepConfigureVNC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	msg := fmt.Sprintf("Looking for available port between %d and %d on %s", config.VNCPortMin, config.VNCPortMax, config.VNCBindAddress)
	ui.Say(msg)
	log.Print(msg)

	var vncPassword string
	var err error
	s.l, err = net.ListenRangeConfig{
		Addr:    config.VNCBindAddress,
		Min:     config.VNCPortMin,
		Max:     config.VNCPortMax,
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		err := fmt.Errorf("Error finding VNC port: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.l.Listener.Close() // free port, but don't unlock lock file
	vncPort := s.l.Port

	if config.VNCUsePassword {
		vncPassword = VNCPassword()
	} else {
		vncPassword = ""
	}

	log.Printf("Found available VNC port: %d on IP: %s", vncPort, config.VNCBindAddress)
	state.Put("vnc_port", vncPort)
	state.Put("vnc_password", vncPassword)

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
