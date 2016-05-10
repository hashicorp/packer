package openstack

import (
	"crypto/rsa"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"golang.org/x/crypto/ssh"
)

// StepGetPassword reads the password from a booted OpenStack server and sets
// it on the WinRM config.
type StepGetPassword struct {
	Debug bool
	Comm  *communicator.Config
}

func (s *StepGetPassword) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// Skip if we're not using winrm
	if s.Comm.Type != "winrm" {
		log.Printf("[INFO] Not using winrm communicator, skipping get password...")
		return multistep.ActionContinue
	}

	// If we already have a password, skip it
	if s.Comm.WinRMPassword != "" {
		ui.Say("Skipping waiting for password since WinRM password set...")
		return multistep.ActionContinue
	}

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Waiting for password since WinRM password is not set...")
	server := state.Get("server").(*servers.Server)
	var password string

	privateKey, err := ssh.ParseRawPrivateKey([]byte(state.Get("privateKey").(string)))
	if err != nil {
		err = fmt.Errorf("Error parsing private key: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	for ; password == "" && err == nil; password, err = servers.GetPassword(computeClient, server.ID).ExtractPassword(privateKey.(*rsa.PrivateKey)) {

		// Check for an interrupt in between attempts.
		if _, ok := state.GetOk(multistep.StateCancelled); ok {
			return multistep.ActionHalt
		}

		log.Printf("Retrying to get a administrator password evry 5 seconds.")
		time.Sleep(5 * time.Second)
	}

	ui.Message(fmt.Sprintf("Password retrieved!"))
	s.Comm.WinRMPassword = password

	// In debug-mode, we output the password
	if s.Debug {
		ui.Message(fmt.Sprintf(
			"Password (since debug is enabled) \"%s\"", s.Comm.WinRMPassword))
	}

	return multistep.ActionContinue
}

func (s *StepGetPassword) Cleanup(multistep.StateBag) {}
