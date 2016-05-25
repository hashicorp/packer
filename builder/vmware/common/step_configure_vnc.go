package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"

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
	VNCBindAddress string
	VNCPortMin     uint
	VNCPortMax     uint
}

type VNCAddressFinder interface {
	VNCAddress(string, uint, uint) (string, uint, error)
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

func (s *StepConfigureVNC) Run(state multistep.StateBag) multistep.StepAction {
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

	log.Printf("Found available VNC port: %d", vncPort)

	vmxData := ParseVMX(string(vmxBytes))
	vmxData["remotedisplay.vnc.enabled"] = "TRUE"
	vmxData["remotedisplay.vnc.port"] = fmt.Sprintf("%d", vncPort)
	vmxData["remotedisplay.vnc.ip"] = fmt.Sprintf("%s", vncBindAddress)

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error writing VMX data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vnc_port", vncPort)
	state.Put("vnc_ip", vncBindAddress)

	return multistep.ActionContinue
}

func (StepConfigureVNC) Cleanup(multistep.StateBag) {
}
