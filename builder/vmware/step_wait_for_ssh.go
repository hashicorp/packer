package vmware

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"net"
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
//   communicator packer.Communicator
type stepWaitForSSH struct{}

func (s *stepWaitForSSH) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	ui.Say("Waiting for SSH to become available...")
	var comm packer.Communicator
	for {
		time.Sleep(5 * time.Second)

		// First we wait for the IP to become available...
		log.Println("Lookup up IP information...")
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

		// Attempt to connect to SSH port
		nc, err := net.Dial("tcp", fmt.Sprintf("%s:22", ip))
		if err != nil {
			log.Printf("TCP connection to SSH ip/port failed: %s", err)
			continue
		}

		// Then we attempt to connect via SSH
		sshConfig := &gossh.ClientConfig{
			User: config.SSHUser,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthPassword(ssh.Password(config.SSHPassword)),
			},
		}

		comm, err = ssh.New(nc, sshConfig)
		if err != nil {
			ui.Error(fmt.Sprintf("Error connecting via SSH: %s", err))
			return multistep.ActionHalt
		}

		ui.Say("Connected via SSH!")
		break
	}

	state["communicator"] = comm

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
