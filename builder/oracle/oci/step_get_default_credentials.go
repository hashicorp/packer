package oci

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepGetDefaultCredentials struct {
	Debug     bool
	Comm      *communicator.Config
	BuildName string
}

func (s *stepGetDefaultCredentials) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		driver = state.Get("driver").(*driverOCI)
		ui     = state.Get("ui").(packersdk.Ui)
		id     = state.Get("instance_id").(string)
	)

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

	username, password, err := driver.GetInstanceInitialCredentials(ctx, id)
	if err != nil {
		err = fmt.Errorf("Error getting instance's credentials: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.Comm.WinRMPassword = password
	s.Comm.WinRMUser = username

	if s.Debug {
		ui.Message(fmt.Sprintf(
			"[DEBUG] (OCI default credentials): Credentials (since debug is enabled): %s", password))
	}

	// store so that we can access this later during provisioning
	state.Put("winrm_password", s.Comm.WinRMPassword)
	packersdk.LogSecretFilter.Set(s.Comm.WinRMPassword)
	return multistep.ActionContinue
}

func (s *stepGetDefaultCredentials) Cleanup(state multistep.StateBag) {
	// no cleanup
}
