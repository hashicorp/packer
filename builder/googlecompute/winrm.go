package googlecompute

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/helper/communicator"
)

// sshConfig returns the ssh configuration.
func winrmConfig(state multistep.StateBag) (*communicator.WinRMConfig, error) {
	config := state.Get("config").(*Config)
	password := state.Get("winrm_password").(string)

	return &communicator.WinRMConfig{
		Username: config.Comm.WinRMUser,
		Password: password,
	}, nil
}
