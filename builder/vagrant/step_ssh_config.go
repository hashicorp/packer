package vagrant

import (
	"context"
	"log"
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

func (s *StepSSHConfig) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	config := state.Get("config").(*Config)

	box := "source"
	if s.GlobalID != "" {
		box = s.GlobalID
	}
	sshConfig, err := driver.SSHConfig(box)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	config.Comm.SSHHost = sshConfig.Hostname
	port, err := strconv.Atoi(sshConfig.Port)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	config.Comm.SSHPort = port

	if config.Comm.SSHUsername != "" {
		// If user has set the username within the communicator, use the
		// auth provided there.
		return multistep.ActionContinue
	}
	log.Printf("identity file is %s", sshConfig.IdentityFile)
	log.Printf("Removing quotes from identity file")
	sshConfig.IdentityFile, err = strconv.Unquote(sshConfig.IdentityFile)
	if err != nil {
		log.Printf("Error unquoting identity file: %s", err)
	}
	config.Comm.SSHPrivateKeyFile = sshConfig.IdentityFile
	config.Comm.SSHUsername = sshConfig.User

	return multistep.ActionContinue
}

func (s *StepSSHConfig) Cleanup(state multistep.StateBag) {
}
