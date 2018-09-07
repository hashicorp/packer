package ecs

import (
	"time"

	"github.com/hashicorp/packer/helper/multistep"
)

var (
	// modified in tests
	sshHostSleepDuration = time.Second
)

type alicloudSSHHelper interface {
}

// SSHHost returns a function that can be given to the SSH communicator
func SSHHost(e alicloudSSHHelper, private bool) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		ipAddress := state.Get("ipaddress").(string)
		return ipAddress, nil
	}
}
