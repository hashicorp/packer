package vmware

import (
	"errors"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// This step waits for SSH to become available and establishes an SSH
// connection.
//
// Uses:
//   config *config
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type stepWaitForSSH struct{}

func (s *stepWaitForSSH) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	ui.Say("Waiting for SSH to become available...")
	for {
		time.Sleep(5 * time.Second)

		log.Println("Lookup up IP information...")
		// First we wait for the IP to become available...
		ipLookup, err := s.dhcpLeaseLookup(vmxPath)
		if err != nil {
			log.Printf("Can't lookup via DHCP lease: %s", err)
		}

		ip, err := ipLookup.GuestIP()
		if err != nil {
			log.Printf("IP lookup failed: %s", err)
			continue
		}

		log.Printf("Detected IP: %s", ip)
	}

	return multistep.ActionContinue
}

func (s *stepWaitForSSH) Cleanup(map[string]interface{}) {}

// Reads the network information for lookup via DHCP.
func (s *stepWaitForSSH) dhcpLeaseLookup(vmxPath string) (GuestIPFinder, error) {
	f, err := os.Open(vmxPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vmxBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	vmxData := ParseVMX(string(vmxBytes))

	var ok bool
	macAddress := ""
	if macAddress, ok = vmxData["ethernet0.address"]; !ok || macAddress == "" {
		if macAddress, ok = vmxData["ethernet0.generatedAddress"]; !ok || macAddress == "" {
			return nil, errors.New("couldn't find MAC address in VMX")
		}
	}

	return &DHCPLeaseGuestLookup{"vmnet8", macAddress}, nil
}
