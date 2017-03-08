package common

import (
	"log"
	"math/rand"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step configures the VM to enable the VNC server.
//
// Uses:
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   vnc_port uint - The port that VNC is configured to listen on.
type StepConfigureVNC struct {
	VNCBindAddress     string
	VNCPortMin         uint
	VNCPortMax         uint
	VNCDisablePassword bool
}

func VNCPassword(skipPassword bool) string {
	if skipPassword {
		return ""
	}
	length := int(8)

	charSet := []byte("012345689abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	charSetLength := len(charSet)

	password := make([]byte, length)

	for i := 0; i < length; i++ {
		password[i] = charSet[rand.Intn(charSetLength)]
	}

	return string(password)
}

func (s *StepConfigureVNC) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	vncPassword := VNCPassword(s.VNCDisablePassword)
	log.Printf("Enabling VNC using an available port between %d and %d", s.VNCPortMin, s.VNCPortMax)
	vncBindAddress, vncPort, err := driver.VNCEnable(vncPassword, s.VNCPortMin, s.VNCPortMax)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Found available VNC port: %d", vncPort)

	state.Put("vnc_port", vncPort)
	state.Put("vnc_ip", vncBindAddress)
	state.Put("vnc_password", vncPassword)

	return multistep.ActionContinue
}

func (StepConfigureVNC) Cleanup(multistep.StateBag) {
}
