package vagrant

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func CommHost() func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		config := state.Get("config").(*Config)
		return config.Comm.SSHHost, nil
	}
}

func SSHPort() func(multistep.StateBag) (uint, error) {
	return func(state multistep.StateBag) (uint, error) {
		config := state.Get("config").(*Config)
		return config.Comm.SSHPort, nil
	}
}
