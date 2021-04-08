package googlecompute

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	metadata "cloud.google.com/go/compute/metadata"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"google.golang.org/api/oauth2/v2"
)

// StepImportOSLoginSSHKey imports a temporary SSH key pair into a GCE login profile.
type StepImportOSLoginSSHKey struct {
	Debug         bool
	TokeninfoFunc func(context.Context) (*oauth2.Tokeninfo, error)
	accountEmail  string
}

// Run executes the Packer build step that generates SSH key pairs.
// The key pairs are added to the ssh config
func (s *StepImportOSLoginSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if !config.UseOSLogin {
		return multistep.ActionContinue
	}

	// If no public key information is available chances are that a private key was provided
	//  or that the user is using a SSH agent for authentication.
	if config.Comm.SSHPublicKey == nil {
		ui.Say("No public SSH key found; skipping SSH public key import for OSLogin...")
		return multistep.ActionContinue
	}

	// Are we running packer on a GCE ?
	s.accountEmail = getGCEUser()

	if s.TokeninfoFunc == nil && s.accountEmail == "" {
		s.TokeninfoFunc = tokeninfo
	}

	ui.Say("Importing SSH public key for OSLogin...")
	// Generate SHA256 fingerprint of SSH public key
	// Put it into state to clean up later
	sha256sum := sha256.Sum256(config.Comm.SSHPublicKey)
	state.Put("ssh_key_public_sha256", hex.EncodeToString(sha256sum[:]))

	if config.account != nil && s.accountEmail == "" {
		s.accountEmail = config.account.jwt.Email
	}

	if s.accountEmail == "" {
		info, err := s.TokeninfoFunc(ctx)
		if err != nil {
			err := fmt.Errorf("Error obtaining token information needed for OSLogin: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.accountEmail = info.Email
	}

	if s.accountEmail == "" {
		err := fmt.Errorf("All options for deriving the OSLogin user have been exhausted")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
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
	ui := state.Get("ui").(packersdk.Ui)

	if !config.UseOSLogin {
		return
	}

	fingerprint, ok := state.Get("ssh_key_public_sha256").(string)
	if !ok || fingerprint == "" {
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

func tokeninfo(ctx context.Context) (*oauth2.Tokeninfo, error) {
	svc, err := oauth2.NewService(ctx)
	if err != nil {
		err := fmt.Errorf("Error initializing oauth service needed for OSLogin: %s", err)
		return nil, err
	}

	return svc.Tokeninfo().Context(ctx).Do()
}

// getGCEUser determines if we're running packer on a GCE, and if we are, gets the associated service account email for subsequent use with OSLogin.
// There are cases where we are running on a GCE, but the GCP metadata server isn't accessible. GitLab docker-engine runners are an edge case example of this.
// It makes little sense to run packer on GCP in this way, however, we defensively timeout in those cases, rather than abort.
func getGCEUser() string {

	metadataCheckTimeout := 5 * time.Second
	metadataCheckChl := make(chan string, 1)

	go func() {
		if metadata.OnGCE() {
			GCEUser, _ := metadata.NewClient(&http.Client{}).Email("")
			metadataCheckChl <- GCEUser
		}
	}()

	select {
	case thisGCEUser := <-metadataCheckChl:
		log.Printf("[INFO] OSLogin: GCE service account %s will be used for identity", thisGCEUser)
		return thisGCEUser
	case <-time.After(metadataCheckTimeout):
		log.Printf("[INFO] OSLogin: Could not derive a GCE service account from google metadata server after %s", metadataCheckTimeout)
		return ""
	}
}
