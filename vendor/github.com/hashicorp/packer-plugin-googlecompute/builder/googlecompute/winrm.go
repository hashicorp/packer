package googlecompute

import (
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// winrmConfig returns the WinRM configuration.
func winrmConfig(state multistep.StateBag) (*communicator.WinRMConfig, error) {
	config := state.Get("config").(*Config)
	password := state.Get("winrm_password").(string)

	return &communicator.WinRMConfig{
		Username: config.Comm.WinRMUser,
		Password: password,
	}, nil
}
