package vagrant

import (
	"context"
	"strconv"

	"github.com/hashicorp/packer/helper/multistep"
)

// Vagrant already sets up ssh on the guests; our job is to find out what
// it did. We can do that with the ssh-config command.  Example output:

// $ vagrant ssh-config
// Host default
//   HostName 172.16.41.194
//   User vagrant
//   Port 22
//   UserKnownHostsFile /dev/null
//   StrictHostKeyChecking no
//   PasswordAuthentication no
//   IdentityFile /Users/mmarsh/Projects/vagrant-boxes/ubuntu/.vagrant/machines/default/vmware_fusion/private_key
//   IdentitiesOnly yes
//   LogLevel FATAL

type StepSSHConfig struct {
	GlobalID string
}

func (s *StepSSHConfig) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	config := state.Get("config").(*Config)

	sshConfig, err := driver.SSHConfig(s.GlobalID)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	config.Comm.SSHPrivateKeyFile = sshConfig.IdentityFile
	config.Comm.SSHUsername = sshConfig.User
	config.Comm.SSHHost = sshConfig.Hostname
	port, err := strconv.Atoi(sshConfig.Port)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	config.Comm.SSHPort = port

	return multistep.ActionContinue
}

func (s *StepSSHConfig) Cleanup(state multistep.StateBag) {
}
