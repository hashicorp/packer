package googlecompute

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"google.golang.org/api/oauth2/v2"
)

// StepCreateSSHKey represents a Packer build step that generates SSH key pairs.
type StepImportOSLoginSSHKey struct {
	Debug        bool
	accountEmail string
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

	// If no public key information is available chances are that a private key was provided
	//  or that the user is using a SSH agent for authentication.
	if config.Comm.SSHPublicKey == nil {
		ui.Say("No public SSH key found; skipping SSH public key import for OSLogin...")
		return multistep.ActionContinue
	}

	ui.Say("Importing SSH public key for OSLogin...")
	// Generate SHA256 fingerprint of SSH public key
	// Put it into state to clean up later
	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	state.Put("ssh_key_public_sha256", hex.EncodeToString(sha256sum[:]))

	if config.account != nil {
		s.accountEmail = config.account.Email
	}

	if s.accountEmail == "" {
		svc, err := oauth2.NewService(ctx)
		if err != nil {
			err := fmt.Errorf("Error initializing oauth service needed for OsLogin: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		info, err := svc.Tokeninfo().Context(ctx).Do()
		if err != nil {
			err := fmt.Errorf("Error obtaining token information needed for OsLogin: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		s.accountEmail = info.Email
	}

	loginProfile, err := driver.ImportOSLoginSSHKey(s.accountEmail, string(config.Comm.SSHPublicKey))
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

	if !config.UseOSLogin {
		return
	}

	fingerprint := state.Get("ssh_key_public_sha256").(string)
	if fingerprint == "" {
		return
	}

	ui.Say("Deleting SSH public key for OSLogin...")
	err := driver.DeleteOSLoginSSHKey(s.accountEmail, fingerprint)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting SSH public key for OSLogin. Please delete it manually.\n\nError: %s", err))
		return
	}

	ui.Message("SSH public key for OSLogin has been deleted!")
}
