package common

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
//   ui     packersdk.Ui
//   vmx_path string
//
// Produces:
//   vnc_port int - The port that VNC is configured to listen on.
type StepConfigureVNC struct {
	Enabled            bool
	VNCBindAddress     string
	VNCPortMin         int
	VNCPortMax         int
	VNCDisablePassword bool

	l *net.Listener
}

type VNCAddressFinder interface {
	VNCAddress(context.Context, string, int, int) (string, int, error)

	// UpdateVMX, sets driver specific VNC values to VMX data.
	UpdateVMX(vncAddress, vncPassword string, vncPort int, vmxData map[string]string)
}

func (s *StepConfigureVNC) VNCAddress(ctx context.Context, vncBindAddress string, portMin, portMax int) (string, int, error) {
	var err error
	s.l, err = net.ListenRangeConfig{
		Addr:    s.VNCBindAddress,
		Min:     s.VNCPortMin,
		Max:     s.VNCPortMax,
		Network: "tcp",
	}.Listen(ctx)
	if err != nil {
		return "", 0, err
	}

	s.l.Listener.Close() // free port, but don't unlock lock file
	return s.l.Address, s.l.Port, nil
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
	ui := state.Get("ui").(packersdk.Ui)
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
	vncBindAddress, vncPort, err := vncFinder.VNCAddress(ctx, s.VNCBindAddress, s.VNCPortMin, s.VNCPortMax)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vncPassword := VNCPassword(s.VNCDisablePassword)

	log.Printf("Found available VNC port: %s:%d", vncBindAddress, vncPort)

	vncFinder.UpdateVMX(vncBindAddress, vncPassword, vncPort, vmxData)

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error writing VMX data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vnc_port", vncPort)
	state.Put("vnc_ip", vncBindAddress)
	state.Put("vnc_password", vncPassword)

	return multistep.ActionContinue
}

func (*StepConfigureVNC) UpdateVMX(address, password string, port int, data map[string]string) {
	data["remotedisplay.vnc.enabled"] = "TRUE"
	data["remotedisplay.vnc.port"] = fmt.Sprintf("%d", port)
	data["remotedisplay.vnc.ip"] = address
	if len(password) > 0 {
		data["remotedisplay.vnc.password"] = password
	}
}

func (s *StepConfigureVNC) Cleanup(multistep.StateBag) {
	if s.l != nil {
		if err := s.l.Close(); err != nil {
			log.Printf("failed to unlock port lockfile: %v", err)
		}
	}
}
