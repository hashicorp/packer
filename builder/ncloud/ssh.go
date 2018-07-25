package ncloud

import (
	packerssh "github.com/hashicorp/packer/communicator/ssh"
	"github.com/hashicorp/packer/helper/multistep"
	"golang.org/x/crypto/ssh"
)

// SSHConfig returns a function that can be used for the SSH communicator
// config for connecting to the specified host via SSH
func SSHConfig(username string) func(multistep.StateBag) (*ssh.ClientConfig, error) {
	return func(state multistep.StateBag) (*ssh.ClientConfig, error) {
		password := state.Get("Password").(string)

		return &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
				ssh.KeyboardInteractive(
					packerssh.PasswordKeyboardInteractive(password)),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}, nil
	}
}
