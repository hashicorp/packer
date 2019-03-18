package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
	Enabled            bool
	VNCBindAddress     string
	VNCPortMin         uint
	VNCPortMax         uint
	VNCDisablePassword bool

	l *net.Listener
}

type VNCAddressFinder interface {
	// UpdateVMX, sets driver specific VNC values to VMX data.
	UpdateVMX(vncAddress, vncPassword string, vncPort uint, vmxData map[string]string)
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

func (s *StepConfigureVNC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.Enabled {
		log.Println("Skipping VNC configuration step...")
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	vmxData, err := ReadVMX(vmxPath)
	if err != nil {
		err := fmt.Errorf("Error reading VMX file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var vncFinder VNCAddressFinder
	if finder, ok := driver.(VNCAddressFinder); ok {
		vncFinder = finder
	} else {
		vncFinder = s
	}

	log.Printf("Looking for available port between %d and %d", s.VNCPortMin, s.VNCPortMax)
	s.l, err = net.ListenRangeConfig{
		Addr:    s.VNCBindAddress,
		Min:     s.VNCPortMin,
		Max:     s.VNCPortMax,
		Network: "tcp",
	}.Listen(ctx)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.l.Listener.Close() // free port, but don't unlock lock file

	vncPassword := VNCPassword(s.VNCDisablePassword)

	log.Printf("Found available VNC port: %v", s.l)

	vncFinder.UpdateVMX(s.l.Address, vncPassword, s.l.Port, vmxData)

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error writing VMX data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vnc_port", s.l.Port)
	state.Put("vnc_ip", s.l.Address)
	state.Put("vnc_password", vncPassword)

	return multistep.ActionContinue
}

func (StepConfigureVNC) UpdateVMX(address, password string, port uint, data map[string]string) {
	data["remotedisplay.vnc.enabled"] = "TRUE"
	data["remotedisplay.vnc.port"] = fmt.Sprintf("%d", port)
	data["remotedisplay.vnc.ip"] = address
	if len(password) > 0 {
		data["remotedisplay.vnc.password"] = password
	}
}

func (s *StepConfigureVNC) Cleanup(multistep.StateBag) {
	if err := s.l.Close(); err != nil {
		log.Printf("failed to unlock port lockfile: %v", err)
	}
}
