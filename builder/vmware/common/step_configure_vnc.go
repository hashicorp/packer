package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"

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
}

type VNCAddressFinder interface {
	VNCAddress(string, uint, uint) (string, uint, error)

	// UpdateVMX, sets driver specific VNC values to VMX data.
	UpdateVMX(vncAddress, vncPassword string, vncPort uint, vmxData map[string]string)
}

func (StepConfigureVNC) VNCAddress(vncBindAddress string, portMin, portMax uint) (string, uint, error) {
	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	var vncPort uint
	portRange := int(portMax - portMin)
	for {
		if portRange > 0 {
			vncPort = uint(rand.Intn(portRange)) + portMin
		} else {
			vncPort = portMin
		}

		log.Printf("Trying port: %d", vncPort)
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", vncBindAddress, vncPort))
		if err == nil {
			defer l.Close()
			break
		}
	}
	return vncBindAddress, vncPort, nil
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
	if !s.Enabled {
		log.Println("Skipping VNC configuration step...")
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	f, err := os.Open(vmxPath)
	if err != nil {
		err := fmt.Errorf("Error reading VMX data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vmxBytes, err := ioutil.ReadAll(f)
	if err != nil {
		err := fmt.Errorf("Error reading VMX data: %s", err)
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
	vncBindAddress, vncPort, err := vncFinder.VNCAddress(s.VNCBindAddress, s.VNCPortMin, s.VNCPortMax)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vncPassword := VNCPassword(s.VNCDisablePassword)

	log.Printf("Found available VNC port: %d", vncPort)

	vmxData := ParseVMX(string(vmxBytes))
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

func (StepConfigureVNC) UpdateVMX(address, password string, port uint, data map[string]string) {
	data["remotedisplay.vnc.enabled"] = "TRUE"
	data["remotedisplay.vnc.port"] = fmt.Sprintf("%d", port)
	data["remotedisplay.vnc.ip"] = address
	if len(password) > 0 {
		data["remotedisplay.vnc.password"] = password
	}
}

func (StepConfigureVNC) Cleanup(multistep.StateBag) {
}
