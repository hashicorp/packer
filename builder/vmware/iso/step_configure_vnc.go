package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
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

type VNCAddressFinder interface {
	VNCAddress(uint, uint) (string, uint)
}

func (stepConfigureVNC) VNCAddress(portMin, portMax uint) (string, uint) {
	// Find an open VNC port. Note that this can still fail later on
	// because we have to release the port at some point. But this does its
	// best.
	var vncPort uint
	portRange := int(portMax - portMin)
	for {
		vncPort = uint(rand.Intn(portRange)) + portMin
		log.Printf("Trying port: %d", vncPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", vncPort))
		if err == nil {
			defer l.Close()
			break
		}
	}
	return "127.0.0.1", vncPort
}

func (s *stepConfigureVNC) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(vmwcommon.Driver)
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
	log.Printf("Looking for available port between %d and %d", config.VNCPortMin, config.VNCPortMax)
	vncIp, vncPort := vncFinder.VNCAddress(config.VNCPortMin, config.VNCPortMax)
	if vncPort == 0 {
		err := fmt.Errorf("Unable to find available VNC port between %d and %d",
			config.VNCPortMin, config.VNCPortMax)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Found available VNC port: %d", vncPort)

	vmxData := vmwcommon.ParseVMX(string(vmxBytes))
	vmxData["remotedisplay.vnc.enabled"] = "TRUE"
	vmxData["remotedisplay.vnc.port"] = fmt.Sprintf("%d", vncPort)

	if err := vmwcommon.WriteVMX(vmxPath, vmxData); err != nil {
		err := fmt.Errorf("Error writing VMX data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vnc_port", vncPort)
	state.Put("vnc_ip", vncIp)

	return multistep.ActionContinue
}

func (stepConfigureVNC) Cleanup(multistep.StateBag) {
}
