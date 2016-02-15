package openstack

import (
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

// StepGetPassword reads the password from a booted OpenStack server and sets
// it on the WinRM config.
type StepGetPassword struct {
	Debug   bool
	Comm    *communicator.Config
}

func (s *StepGetPassword) Run(state multistep.StateBag) multistep.StepAction {
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

	server := state.Get("server").(*servers.Server)
	password := server.AdminPass
	// TODO: Are there relevant error cases here? If we got here and AdminPass
	// is "", the communicator password is empty string anyway? Log a warning?

	ui.Message(fmt.Sprintf("Password retrieved!"))
	s.Comm.WinRMPassword = password

	// In debug-mode, we output the password
	if s.Debug {
		ui.Message(fmt.Sprintf(
			"Password (since debug is enabled): %s", s.Comm.WinRMPassword))
	}

	return multistep.ActionContinue
}

func (s *StepGetPassword) Cleanup(multistep.StateBag) {}
