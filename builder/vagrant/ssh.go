package vagrant

import (
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

func CommHost() func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		config := state.Get("config").(*Config)
		return config.Comm.SSHHost, nil
	}
}

func SSHPort() func(multistep.StateBag) (int, error) {
	return func(state multistep.StateBag) (int, error) {
		config := state.Get("config").(*Config)
		return config.Comm.SSHPort, nil
	}
}
