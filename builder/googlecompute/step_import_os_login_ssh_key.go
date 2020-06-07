package googlecompute

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepCreateSSHKey represents a Packer build step that generates SSH key pairs.
type StepImportOSLoginSSHKey struct {
	Debug bool
}

// Run executes the Packer build step that generates SSH key pairs.
// The key pairs are added to the ssh config
func (s *StepImportOSLoginSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if !config.UseOSLogin {
		return multistep.ActionContinue
	}

	ui.Say("Importing SSH public key for OSLogin...")

	// Generate SHA256 fingerprint of SSH public key
	// Put it into state to clean up later
	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	state.Put("ssh_key_public_sha256", hex.EncodeToString(sha256sum[:]))

	// TODO: @cpwc need to obtain user's email
	if config.account == nil {
		err := fmt.Errorf("Error importing SSH public key for OSLogin: need user's ID/email")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	loginProfile, err := driver.ImportOSLoginSSHKey(config.account.Email, string(config.Comm.SSHPublicKey))
	if err != nil {
		err := fmt.Errorf("Error importing SSH public key for OSLogin: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Replacing `SSHUsername` as the username have to be from OSLogin
	if len(loginProfile.PosixAccounts) == 0 {
		err := fmt.Errorf("Error importing SSH public key for OSLogin: no PosixAccounts available")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Let's obtain the `Primary` account username
	ui.Say("Obtaining SSH Username for OSLogin...")
	var username string
	for _, account := range loginProfile.PosixAccounts {
		if account.Primary {
			username = account.Username
			break
		}
	}

	if s.Debug {
		ui.Message(fmt.Sprintf("ssh_username: %s", username))
	}
	config.Comm.SSHUsername = username

	return multistep.ActionContinue
}

// Cleanup the SSH Key that we added to the POSIX account
func (s *StepImportOSLoginSSHKey) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting SSH public key for OSLogin...")

	// TODO: @cpwc need to obtain user's email
	if config.account == nil {
		err := fmt.Errorf("Error deleting SSH public key for OSLogin: need user's ID/email")
		state.Put("error", err)
		ui.Error(err.Error())
		return
	}

	fingerprint := state.Get("ssh_key_public_sha256").(string)
	err := driver.DeleteOSLoginSSHKey(config.account.Email, fingerprint)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting SSH public key for OSLogin. Please delete it manually.\n\nError: %s", err))
		return
	}

	ui.Message("SSH public key for OSLogin has been deleted!")
}
