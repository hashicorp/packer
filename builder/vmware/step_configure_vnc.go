package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
)

// This step configures the VM to enable the VNC server.
//
// Uses:
//   config *config
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   vnc_port uint - The port that VNC is configured to listen on.
type stepConfigureVNC struct{}

func (stepConfigureVNC) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
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

	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	log.Printf("Looking for available port between %d and %d", config.VNCPortMin, config.VNCPortMax)
	var vncPort uint
	portRange := int(config.VNCPortMax - config.VNCPortMin)
	for {
		vncPort = uint(rand.Intn(portRange)) + config.VNCPortMin
		log.Printf("Trying port: %d", vncPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", vncPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	log.Printf("Found available VNC port: %d", vncPort)

	vmxData := ParseVMX(string(vmxBytes))
	vmxData["RemoteDisplay.vnc.enabled"] = "TRUE"
	vmxData["RemoteDisplay.vnc.port"] = fmt.Sprintf("%d", vncPort)

	if err := WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error writing VMX data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vnc_port", vncPort)

	return multistep.ActionContinue
}

func (stepConfigureVNC) Cleanup(multistep.StateBag) {
}
